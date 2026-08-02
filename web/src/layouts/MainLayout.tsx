import { useMemo, useState } from 'react'
import { Outlet, useLocation, useNavigate } from 'react-router-dom'
import { Avatar, Dropdown, Layout, Menu, theme } from 'antd'
import {
  DashboardOutlined, SettingOutlined, UserOutlined, TeamOutlined, MenuOutlined,
  ApartmentOutlined, BookOutlined, ToolOutlined, ClockCircleOutlined, FileTextOutlined,
  LoginOutlined, FolderOutlined, MonitorOutlined, LogoutOutlined,
  MenuFoldOutlined, MenuUnfoldOutlined,
} from '@ant-design/icons'
import type { MenuProps } from 'antd'
import { useAuthStore } from '@/store/auth'
import type { Menu as MenuModel } from '@/types'

const { Header, Sider, Content } = Layout

// 图标名到组件的映射：后端菜单 icon 字段用名称，前端在此解析。
const iconMap: Record<string, React.ReactNode> = {
  setting: <SettingOutlined />,
  user: <UserOutlined />,
  team: <TeamOutlined />,
  menu: <MenuOutlined />,
  dept: <ApartmentOutlined />,
  dict: <BookOutlined />,
  config: <ToolOutlined />,
  job: <ClockCircleOutlined />,
  'oper-log': <FileTextOutlined />,
  'login-log': <LoginOutlined />,
  file: <FolderOutlined />,
  monitor: <MonitorOutlined />,
}

// 按路径推断默认图标（未配置 icon 时）。
function pickIcon(m: MenuModel): React.ReactNode {
  if (m.icon && iconMap[m.icon]) return iconMap[m.icon]
  return iconMap[m.path] || <FileTextOutlined />
}

// buildMenuItems 将后端菜单树转为 antd Menu 数据；只保留目录/菜单（type 1/2）。
// key 使用完整路由路径（父路径 + 自身路径）。
function buildMenuItems(menus: MenuModel[], parentPath: string): Required<MenuProps>['items'] {
  return menus
    .filter((m) => m.type !== 3 && m.visible === 1)
    .map((m) => {
      const fullPath = `${parentPath}/${m.path}`.replace(/\/+/g, '/')
      const children = m.children ? buildMenuItems(m.children, fullPath) : []
      return {
        key: fullPath,
        icon: pickIcon(m),
        label: m.title,
        children: children && children.length > 0 ? children : undefined,
      }
    })
}

export default function MainLayout() {
  const navigate = useNavigate()
  const location = useLocation()
  const [collapsed, setCollapsed] = useState(false)
  const { user, menus, logout } = useAuthStore()
  const { token } = theme.useToken()

  const items = useMemo<Required<MenuProps>['items']>(() => {
    const dashboard = { key: '/dashboard', icon: <DashboardOutlined />, label: '工作台' }
    return [dashboard, ...buildMenuItems(menus, '')]
  }, [menus])

  // 展开当前路径的父级菜单。
  const openKeys = useMemo(() => {
    const segs = location.pathname.split('/').filter(Boolean)
    const keys: string[] = []
    let acc = ''
    for (const s of segs) {
      acc += `/${s}`
      keys.push(acc)
    }
    return keys
  }, [location.pathname])

  const userMenu: MenuProps['items'] = [
    { key: 'logout', icon: <LogoutOutlined />, label: '退出登录' },
  ]

  return (
    <Layout style={{ height: '100vh' }}>
      <Sider collapsible collapsed={collapsed} trigger={null} theme="dark" width={220}>
        <div
          style={{
            height: 48,
            margin: 8,
            color: '#fff',
            fontWeight: 600,
            fontSize: collapsed ? 14 : 16,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            whiteSpace: 'nowrap',
            overflow: 'hidden',
          }}
        >
          {collapsed ? 'GF' : 'go-framework'}
        </div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={[location.pathname]}
          defaultOpenKeys={openKeys}
          items={items}
          onClick={({ key }) => navigate(key)}
        />
      </Sider>
      <Layout>
        <Header style={{ padding: '0 16px', background: token.colorBgContainer, display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <div style={{ cursor: 'pointer', fontSize: 18 }} onClick={() => setCollapsed(!collapsed)}>
            {collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
          </div>
          <Dropdown
            menu={{
              items: userMenu,
              onClick: async ({ key }) => {
                if (key === 'logout') {
                  await logout()
                  navigate('/login', { replace: true })
                }
              },
            }}
          >
            <span style={{ cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 8 }}>
              <Avatar size="small" icon={<UserOutlined />} />
              {user?.nickname || user?.username}
            </span>
          </Dropdown>
        </Header>
        <Content style={{ overflow: 'auto', background: token.colorBgLayout }}>
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  )
}
