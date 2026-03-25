import { NextResponse } from 'next/server';
import { sql } from '@/lib/db';

export async function POST(request: Request) {
  try {
    const { token, machineId } = await request.json();

    if (!token) {
      return NextResponse.json({ valid: false, error: 'Token missing' }, { status: 400 });
    }
    if (!machineId) {
      return NextResponse.json({ valid: false, error: 'Machine ID missing' }, { status: 400 });
    }

    const result = await sql`
      SELECT is_active, bound_devices FROM auth_tokens 
      WHERE token = ${token}
    `;

    if (result.length === 0) {
      return NextResponse.json({ valid: false, error: 'Token not found' }, { status: 404 });
    }

    const isActive = result[0].is_active;
    const boundDevices = result[0].bound_devices || [];
    
    if (!isActive) {
      return NextResponse.json({ valid: false, error: 'Token revoked or inactive' }, { status: 403 });
    }

    // Check device binding
    if (!boundDevices.includes(machineId)) {
      if (boundDevices.length >= 3) {
        return NextResponse.json({ valid: false, error: 'Device limit reached' }, { status: 403 });
      } else {
        // Append device and update DB
        await sql`
          UPDATE auth_tokens
          SET bound_devices = array_append(bound_devices, ${machineId})
          WHERE token = ${token}
        `;
      }
    }

    return NextResponse.json({ valid: true });
  } catch (error) {
    console.error('Failed to verify token:', error);
    return NextResponse.json({ valid: false, error: 'Internal Server Error' }, { status: 500 });
  }
}
