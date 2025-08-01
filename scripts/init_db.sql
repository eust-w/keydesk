-- KeyDesk 数据库初始化脚本

-- 创建管理员用户 (密码: admin123)
-- 注意：在生产环境中应该使用更安全的密码
INSERT INTO users (email, password, role, status, created_at, updated_at) 
VALUES (
    'admin@keydesk.com',
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', -- bcrypt hash of 'admin123'
    'admin',
    'active',
    NOW(),
    NOW()
) ON CONFLICT (email) DO NOTHING;

-- 创建一些示例激活码
INSERT INTO activation_keys (code, type, status, max_uses, used_count, expires_at, created_by, created_at, updated_at)
VALUES 
    ('KD1234567890ABCDEF', 'premium', 'unused', 1, 0, NOW() + INTERVAL '1 year', 'admin@keydesk.com', NOW(), NOW()),
    ('KD2345678901BCDEFG', 'standard', 'unused', 5, 0, NOW() + INTERVAL '6 months', 'admin@keydesk.com', NOW(), NOW()),
    ('KD3456789012CDEFGH', 'trial', 'unused', 1, 0, NOW() + INTERVAL '30 days', 'admin@keydesk.com', NOW(), NOW())
ON CONFLICT (code) DO NOTHING; 