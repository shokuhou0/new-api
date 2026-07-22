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
import { api } from '@/lib/api'

type CanvasHandoffData = {
  ticket: string
  canvas_url: string
  expires_at: number
}

type CanvasHandoffResponse = {
  success: boolean
  code?: string
  message?: string
  data?: CanvasHandoffData
}

export class CanvasHandoffError extends Error {
  readonly code?: string

  constructor(message: string, code?: string) {
    super(message)
    this.name = 'CanvasHandoffError'
    this.code = code
  }
}

export async function createCanvasHandoff(): Promise<CanvasHandoffData> {
  const response = await api.post<CanvasHandoffResponse>(
    '/api/canvas/handoff',
    undefined,
    {
      skipBusinessError: true,
      skipErrorHandler: true,
    }
  )
  const payload = response.data
  if (!payload.success || !payload.data) {
    throw new CanvasHandoffError(
      payload.message || 'Canvas handoff failed',
      payload.code
    )
  }
  return payload.data
}

export function buildCanvasHandoffUrl(data: CanvasHandoffData): string {
  const url = new URL(data.canvas_url)
  url.hash = new URLSearchParams({ newapi_handoff: data.ticket }).toString()
  return url.toString()
}
