import { NextResponse } from 'next/server';
import { sql } from '@/lib/db';

export async function POST(request: Request) {
  try {
    const { token } = await request.json();

    if (!token) {
      return NextResponse.json({ valid: false, error: 'Token missing' }, { status: 400 });
    }

    const result = await sql`
      SELECT is_active FROM auth_tokens 
      WHERE token = ${token}
    `;

    if (result.length === 0) {
      return NextResponse.json({ valid: false, error: 'Token not found' }, { status: 404 });
    }

    const isActive = result[0].is_active;
    
    if (!isActive) {
      return NextResponse.json({ valid: false, error: 'Token revoked or inactive' }, { status: 403 });
    }

    return NextResponse.json({ valid: true });
  } catch (error) {
    console.error('Failed to verify token:', error);
    return NextResponse.json({ valid: false, error: 'Internal Server Error' }, { status: 500 });
  }
}
