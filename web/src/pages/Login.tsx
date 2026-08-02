import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Button, Card, Form, Input, App as AntdApp } from 'antd'
import { LockOutlined, UserOutlined, SafetyCertificateOutlined } from '@ant-design/icons'
import { authApi } from '@/api'
import { tokenStore } from '@/api/request'
import { useAuthStore } from '@/store/auth'

export default function Login() {
  const navigate = useNavigate()
  const { message } = AntdApp.useApp()
  const fetchUserInfo = useAuthStore((s) => s.fetchUserInfo)
  const [loading, setLoading] = useState(false)
  const [captcha, setCaptcha] = useState<{ id: string; img: string; enabled: boolean }>({ id: '', img: '', enabled: true })

  const refreshCaptcha = async () => {
    try {
      const c = await authApi.captcha()
      setCaptcha({ id: c.captchaId, img: c.captchaImg, enabled: c.enabled })
    } catch {
      // 忽略：登录时会给出提示
    }
  }

  useEffect(() => {
    refreshCaptcha()
  }, [])

  const onFinish = async (values: { username: string; password: string; captchaCode?: string }) => {
    setLoading(true)
    try {
      const pair = await authApi.login({
        username: values.username,
        password: values.password,
        captchaId: captcha.id,
        captchaCode: values.captchaCode || '',
      })
      tokenStore.set(pair.accessToken, pair.refreshToken)
      await fetchUserInfo()
      message.success('登录成功')
      navigate('/dashboard', { replace: true })
    } catch {
      refreshCaptcha()
    } finally {
      setLoading(false)
    }
  }

  return (
    <div
      style={{
        height: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        background: 'linear-gradient(135deg, #1677ff 0%, #0958d9 100%)',
      }}
    >
      <Card style={{ width: 380, boxShadow: '0 8px 32px rgba(0,0,0,0.15)' }}>
        <h2 style={{ textAlign: 'center', marginBottom: 8 }}>go-framework</h2>
        <p style={{ textAlign: 'center', color: '#888', marginTop: 0 }}>企业级管理后台</p>
        <Form onFinish={onFinish} size="large" initialValues={{ username: 'admin', password: 'admin123' }}>
          <Form.Item name="username" rules={[{ required: true, message: '请输入用户名' }]}>
            <Input prefix={<UserOutlined />} placeholder="用户名" />
          </Form.Item>
          <Form.Item name="password" rules={[{ required: true, message: '请输入密码' }]}>
            <Input.Password prefix={<LockOutlined />} placeholder="密码" />
          </Form.Item>
          {captcha.enabled && (
            <Form.Item name="captchaCode" rules={[{ required: true, message: '请输入验证码' }]}>
              <Input
                prefix={<SafetyCertificateOutlined />}
                placeholder="验证码"
                addonAfter={
                  captcha.img ? (
                    <img
                      src={captcha.img}
                      alt="验证码"
                      style={{ height: 32, cursor: 'pointer' }}
                      onClick={refreshCaptcha}
                      title="点击刷新"
                    />
                  ) : null
                }
              />
            </Form.Item>
          )}
          <Button type="primary" htmlType="submit" block loading={loading}>
            登录
          </Button>
        </Form>
        <p style={{ textAlign: 'center', color: '#bbb', fontSize: 12, marginTop: 16, marginBottom: 0 }}>
          默认账号 admin / admin123
        </p>
      </Card>
    </div>
  )
}
