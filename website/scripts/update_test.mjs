import { neon } from '@neondatabase/serverless';
import dotenv from 'dotenv';
import path from 'path';

dotenv.config({ path: path.resolve(process.cwd(), '.env.local') });

async function updateToken() {
  const sql = neon(process.env.DATABASE_URL);
  const targetToken = 'TP-E96886E1-2872';

  try {
    // 构造3个伪造的机器特征码，同时把原来的真实机器码覆盖掉（相当于删除老数据）
    const fakeDevices = ['fake-device-id-001', 'fake-device-id-002', 'fake-device-id-003'];
    
    await sql`UPDATE auth_tokens SET bound_devices = ${fakeDevices} WHERE token = ${targetToken}`;
    
    console.log(`✅ 注入完成！Token ${targetToken} 的绑定槽位已被彻底占满。`);
    console.log(`🔒 当前库中的占位设备为:`, fakeDevices);
    console.log(`💡 提示：您现在可以去桌面端再次点击验证拉起，预期会立刻触发满载安全拦截！`);
  } catch (error) {
    console.error('更新测试数据出错:', error);
  }
}

updateToken();
