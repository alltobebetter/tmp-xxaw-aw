import { NextResponse } from 'next/server';

export async function GET() {
  const version = '1.2.0';
  return NextResponse.json({
    version: version,
    forceUpdate: true,
    downloadUrl: `https://public.agentlab.click/TraeProxy_Setup_v${version}.exe`,
    downloads: {
      windows: `https://public.agentlab.click/TraeProxy_Setup_v${version}.exe`,
      macos: `https://public.agentlab.click/TraeProxy_macOS_v${version}.dmg`
    }
  });
}
