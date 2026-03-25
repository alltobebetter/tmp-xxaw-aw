import { NextResponse } from 'next/server';
import { sql } from '@/lib/db';
import { v4 as uuidv4 } from 'uuid';

export async function POST(req: Request) {
  try {
    const body = await req.json().catch(() => ({}));
    const { recaptchaToken } = body;

    if (!recaptchaToken) {
      return NextResponse.json({ success: false, error: '缺少人机验证参数' }, { status: 400 });
    }

    // Verify reCAPTCHA using recaptcha.net to bypass GFW
    const verifyRes = await fetch('https://www.recaptcha.net/recaptcha/api/siteverify', {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: `secret=${process.env.RECAPTCHA_SECRET_KEY}&response=${recaptchaToken}`
    });
    
    const verifyData = await verifyRes.json();
    if (!verifyData.success) {
      console.error('reCAPTCHA failed:', verifyData['error-codes']);
      return NextResponse.json({ success: false, error: '人机验证未通过，请刷新页面重试' }, { status: 400 });
    }

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
