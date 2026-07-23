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
import assert from 'node:assert/strict'
import { describe, test } from 'node:test'

import { resolveCanvasLaunchHref } from '../navigation.ts'

describe('Infinite Canvas navigation', () => {
  test('sends anonymous users directly to the configured canvas URL', () => {
    assert.equal(
      resolveCanvasLaunchHref(false, 'https://canvas.example.com'),
      'https://canvas.example.com'
    )
  })

  test('keeps authenticated users on the handoff route', () => {
    assert.equal(
      resolveCanvasLaunchHref(true, 'https://canvas.example.com'),
      '/canvas-launch'
    )
  })

  test('falls back to the handoff route when the canvas URL is unavailable', () => {
    assert.equal(resolveCanvasLaunchHref(false, '   '), '/canvas-launch')
    assert.equal(resolveCanvasLaunchHref(false, undefined), '/canvas-launch')
  })
})
