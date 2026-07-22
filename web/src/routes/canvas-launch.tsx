/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { useMutation } from '@tanstack/react-query'
import { createFileRoute, Link } from '@tanstack/react-router'
import { AlertCircle, Loader2 } from 'lucide-react'
import { useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'
import {
  buildCanvasHandoffUrl,
  CanvasHandoffError,
  createCanvasHandoff,
} from '@/features/canvas-launch/api'
import { useStatus } from '@/hooks/use-status'
import { useAuthStore } from '@/stores/auth-store'

export const Route = createFileRoute('/canvas-launch')({
  component: CanvasLaunchPage,
})

const errorTranslationKeys: Record<string, string> = {
  CANVAS_CONFIG_INVALID:
    'Infinite Canvas is not configured correctly. Contact an administrator.',
  CANVAS_GROUP_UNAVAILABLE:
    'The Image token group is not available for your account.',
  CANVAS_TOKEN_NOT_FOUND:
    'No usable Image group token was found. Create or enable one and try again.',
}

function CanvasLaunchPage() {
  const { t } = useTranslation()
  const { auth } = useAuthStore()
  const { status, loading: statusLoading } = useStatus()
  const started = useRef(false)
  const [redirectError, setRedirectError] = useState(false)
  const isAuthenticated = Boolean(auth.user && auth.accessToken)
  const handoff = useMutation({
    mutationFn: createCanvasHandoff,
    onSuccess: (data) => {
      window.location.replace(buildCanvasHandoffUrl(data))
    },
  })

  useEffect(() => {
    if (started.current) return
    if (isAuthenticated) {
      started.current = true
      handoff.mutate()
      return
    }
    if (statusLoading) return
    const canvasURL = status?.canvas_url
    if (typeof canvasURL === 'string' && canvasURL.trim()) {
      started.current = true
      window.location.replace(canvasURL)
      return
    }
    setRedirectError(true)
  }, [handoff, isAuthenticated, status, statusLoading])

  const errorKey =
    handoff.error instanceof CanvasHandoffError && handoff.error.code
      ? errorTranslationKeys[handoff.error.code]
      : undefined
  const hasError = handoff.isError || redirectError
  let errorMessage = t('Unable to open Infinite Canvas. Please try again.')
  if (redirectError) {
    errorMessage = t(
      'Infinite Canvas is not configured correctly. Contact an administrator.'
    )
  } else if (errorKey) {
    errorMessage = t(errorKey)
  }

  return (
    <main className='flex min-h-screen items-center justify-center px-4 py-12'>
      <section className='bg-card w-full max-w-md rounded-xl border p-6 text-center shadow-sm'>
        {hasError ? (
          <>
            <div className='bg-destructive/10 text-destructive mx-auto flex size-12 items-center justify-center rounded-full'>
              <AlertCircle className='size-6' aria-hidden='true' />
            </div>
            <h1 className='mt-4 text-lg font-semibold'>
              {t('Unable to open Infinite Canvas. Please try again.')}
            </h1>
            <p className='text-muted-foreground mt-2 text-sm'>{errorMessage}</p>
            <div className='mt-6 flex justify-center gap-2'>
              <Button
                onClick={() => {
                  if (isAuthenticated) {
                    handoff.mutate()
                  } else {
                    window.location.reload()
                  }
                }}
                disabled={handoff.isPending}
              >
                {t('Retry')}
              </Button>
              {isAuthenticated && (
                <Button variant='outline' render={<Link to='/keys' />}>
                  {t('API Keys')}
                </Button>
              )}
            </div>
          </>
        ) : (
          <>
            <Loader2
              className='text-primary mx-auto size-10 animate-spin'
              aria-hidden='true'
            />
            <h1 className='mt-4 text-lg font-semibold'>
              {t('Opening Infinite Canvas...')}
            </h1>
            <p className='text-muted-foreground mt-2 text-sm'>
              {t(
                'Your Image token is being transferred securely. This page will redirect automatically.'
              )}
            </p>
          </>
        )}
      </section>
    </main>
  )
}
