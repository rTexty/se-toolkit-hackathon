import { motion } from 'framer-motion'
import { toast } from 'sonner'
import { Loader2 } from 'lucide-react'
import { useDummyLogin } from '@/api/hooks'
import { Link } from 'react-router-dom'

export default function LoginPage() {
  const loginMutation = useDummyLogin()

  const handleLogin = (role: 'admin' | 'user') => {
    loginMutation.mutate(role, {
      onError: () => {
        toast.error('Login failed. Please try again.')
      },
    })
  }

  return (
    <div className="min-h-screen flex items-center justify-center relative overflow-hidden bg-gradient-to-br from-indigo-100 via-purple-50 to-pink-100">
      <div className="absolute inset-0 overflow-hidden pointer-events-none">
        <div className="absolute -top-40 -right-40 w-80 h-80 bg-purple-300 rounded-full mix-blend-multiply filter blur-xl opacity-70 animate-float" />
        <div className="absolute -bottom-40 -left-40 w-80 h-80 bg-indigo-300 rounded-full mix-blend-multiply filter blur-xl opacity-70 animate-float-delay" />
        <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-80 h-80 bg-pink-300 rounded-full mix-blend-multiply filter blur-xl opacity-70 animate-float-slow" />
      </div>

      <motion.div
        initial={{ opacity: 0, y: 30, scale: 0.95 }}
        animate={{ opacity: 1, y: 0, scale: 1 }}
        transition={{ duration: 0.5, ease: 'easeOut' }}
        className="relative z-10 w-full max-w-md mx-4"
      >
        <div className="glass-card p-8 sm:p-10">
          <div className="text-center mb-8">
            <h1 className="text-3xl font-bold bg-gradient-to-r from-indigo-600 to-purple-600 bg-clip-text text-transparent">
              Room Booking
            </h1>
            <p className="text-muted-foreground mt-2">Select your role to continue</p>
          </div>

          <div className="space-y-4">
            <motion.button
              whileHover={{ scale: 1.02 }}
              whileTap={{ scale: 0.98 }}
              onClick={() => handleLogin('admin')}
              disabled={loginMutation.isPending}
              className="w-full h-14 rounded-xl bg-gradient-to-r from-indigo-600 to-purple-600 text-white font-medium shadow-lg shadow-indigo-500/25 transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
            >
              {loginMutation.isPending ? (
                <Loader2 className="h-5 w-5 animate-spin" />
              ) : null}
              {loginMutation.isPending ? 'Logging in...' : 'Login as Admin'}
            </motion.button>

            <motion.button
              whileHover={{ scale: 1.02 }}
              whileTap={{ scale: 0.98 }}
              onClick={() => handleLogin('user')}
              disabled={loginMutation.isPending}
              className="w-full h-14 rounded-xl border-2 border-indigo-200 bg-white/50 backdrop-blur-sm text-indigo-700 font-medium hover:bg-white/80 hover:border-indigo-300 transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
            >
              {loginMutation.isPending ? (
                <Loader2 className="h-5 w-5 animate-spin" />
              ) : null}
              {loginMutation.isPending ? 'Logging in...' : 'Login as User'}
            </motion.button>
          </div>

          <p className="text-center text-sm text-muted-foreground mt-6">
            Don't have an account?{' '}
            <Link to="/register" className="text-indigo-600 hover:text-indigo-700 font-medium">
              Register
            </Link>
          </p>
        </div>
      </motion.div>
    </div>
  )
}
