import { neon } from '@neondatabase/serverless';
import dotenv from 'dotenv';
import path from 'path';

dotenv.config({ path: path.resolve(process.cwd(), '.env.local') });

async function migrate() {
  if (!process.env.DATABASE_URL) {
    console.error('DATABASE_URL is not set.');
    process.exit(1);
  }

  const sql = neon(process.env.DATABASE_URL);

  try {
    console.log('Adding bound_devices column to auth_tokens table...');
    await sql`ALTER TABLE auth_tokens ADD COLUMN IF NOT EXISTS bound_devices TEXT[] DEFAULT '{}';`;
    console.log('Column added successfully.');
  } catch (error) {
    console.error('Migration failed:', error);
  }
}

migrate();
