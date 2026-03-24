import { NextResponse } from 'next/server';
import { sql } from '@/lib/db';
import { v4 as uuidv4 } from 'uuid';

export async function POST() {
  try {
    const rawUuid = uuidv4();
    // Create a distinctive token format: TP-XXXX-XXXX
    const token = `TP-${rawUuid.split('-')[0]}-${rawUuid.split('-')[1]}`.toUpperCase();
    
    const result = await sql`
      INSERT INTO auth_tokens (token, is_active)
      VALUES (${token}, true)
      RETURNING token, created_at
    `;

    return NextResponse.json({ success: true, data: result[0] });
  } catch (error) {
    console.error('Failed to generate token:', error);
    return NextResponse.json({ success: false, error: 'Internal Server Error' }, { status: 500 });
  }
}
