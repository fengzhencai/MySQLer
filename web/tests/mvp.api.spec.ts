import { test, expect, request } from '@playwright/test';

const BASE_URL = 'http://localhost:8090';

const previewPayload = {
  host: '127.0.0.1',
  port: 3307,
  username: 'root',
  password: 'root123456',
  database: 'test_db',
  table: 'test_users',
  ddl_statement: "ADD COLUMN phone VARCHAR(20) NULL AFTER email",
  params: { chunk_size: 1000, dry_run: true }
};

test.describe('MVP API smoke', () => {
  test('health is ok', async ({ request }) => {
    const res = await request.get(`${BASE_URL}/health`);
    expect(res.status()).toBe(200);
    const body = await res.json();
    expect(body.status).toBe('ok');
  });

  test('preview returns generated command', async ({ request }) => {
    const res = await request.post(`${BASE_URL}/api/v1/mvp/preview`, {
      data: previewPayload,
    });
    expect(res.status()).toBe(200);
    const body = await res.json();
    expect(body.code).toBe(200);
    expect(body.data.generated_command).toContain('pt-online-schema-change');
    expect(body.data.generated_command).toContain('--dry-run');
  });

  test('execute (dry_run) returns exit_code and stdout', async ({ request }) => {
    const payload = { ...previewPayload, params: { ...previewPayload.params, dry_run: true } };
    const res = await request.post(`${BASE_URL}/api/v1/mvp/execute`, {
      data: payload,
    });
    expect(res.status()).toBe(200);
    const body = await res.json();
    expect(body.code).toBe(200);
    // exit_code may be 0 even for dry-run; ensure fields exist
    expect(body.data).toHaveProperty('generated_command');
    expect(body.data).toHaveProperty('exit_code');
    expect(body.data).toHaveProperty('stdout');
  });
});
