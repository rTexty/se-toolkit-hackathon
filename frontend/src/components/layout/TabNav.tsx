import { Link, useLocation } from 'react-router-dom'
import { motion } from 'framer-motion'
import { LogOut, User } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useAuthStore } from '@/store/authStore'

const navItems = [
  { path: '/rooms', label: 'Rooms', roles: ['admin', 'user'] },
  { path: '/my-bookings', label: 'My Bookings', roles: ['admin', 'user'] },
  { path: '/admin', label: 'Admin', roles: ['admin'] },
]

export default function TabNav() {
  const location = useLocation()
  const { user, logout } = useAuthStore()

  const filteredNavItems = navItems.filter((item) =>
    user && item.roles.includes(user.role)
  )

  return (
    <nav className="glass-nav sticky top-0 z-50">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-16">
          <div className="flex items-center gap-1">
            {filteredNavItems.map((item) => {
              const isActive = location.pathname === item.path
              return (
                <Button
                  key={item.path}
                  variant="ghost"
                  asChild
                  className={`relative px-4 py-2 text-sm font-medium transition-colors ${
                    isActive
                      ? 'text-indigo-700'
                      : 'text-muted-foreground hover:text-foreground'
                  }`}
                >
                  <Link to={item.path}>
                    {item.label}
                    {isActive && (
                      <motion.div
                        layoutId="tab-indicator"
                        className="absolute -bottom-px left-2 right-2 h-0.5 bg-gradient-to-r from-indigo-500 to-purple-500 rounded-full"
                      />
                    )}
                  </Link>
                </Button>
              )
            })}
          </div>
          <div className="flex items-center gap-3">
            <div className="flex items-center gap-2 text-sm text-muted-foreground glass px-3 py-1.5 rounded-full">
              <User className="w-3.5 h-3.5" />
              <span className="capitalize font-medium">{user?.role}</span>
            </div>
            <Button
              variant="ghost"
              size="icon"
              onClick={logout}
              className="text-muted-foreground hover:text-rose-600 transition-colors"
            >
              <LogOut className="w-4 h-4" />
            </Button>
          </div>
        </div>
      </div>
    </nav>
  )
}
