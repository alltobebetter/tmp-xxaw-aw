import { neon } from '@neondatabase/serverless';
import dotenv from 'dotenv';
import path from 'path';

dotenv.config({ path: path.resolve(process.cwd(), '.env.local') });

async function checkToken() {
  const sql = neon(process.env.DATABASE_URL);
  const targetToken = 'TP-E96886E1-2872';

  try {
    const result = await sql`SELECT bound_devices FROM auth_tokens WHERE token = ${targetToken}`;
    
    if (result.length === 0) {
      console.log(`⚠️ 在数据库中未找到 Token: ${targetToken}`);
      return;
    }

    const devices = result[0].bound_devices || [];
    console.log(`✅ Token: ${targetToken}`);
    console.log(`💻 绑定的设备数量: ${devices.length} 台`);
    
    if (devices.length > 0) {
      console.log(`🔒 设备特征码列表:`, devices);
    }
  } catch (error) {
    console.error('查询出错:', error);
  }
}

checkToken();
