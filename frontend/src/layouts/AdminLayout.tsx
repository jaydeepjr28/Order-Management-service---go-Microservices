import React, { useState } from 'react';
import { Outlet, Link, useLocation, useNavigate } from 'react-router-dom';
import { useAuthStore } from '@/features/auth/useAuthStore';
import { Button } from '@/components/ui/Button';
import {
  LayoutDashboard,
  Package,
  ShoppingCart,
  BarChart3,
  ArrowLeft,
  LogOut,
  ShieldCheck,
  ShieldX,
  Sun,
  Moon,
  Menu,
  X,
  Sparkles,
} from 'lucide-react';

export function AdminLayout() {
  const location = useLocation();
  const navigate = useNavigate();
  const { isAuthenticated, isAdmin, customerID, logout } = useAuthStore();
  const [isSidebarOpen, setIsSidebarOpen] = useState(false);
  const [isDark, setIsDark] = useState(() =>
    document.documentElement.classList.contains('dark')
  );

  const toggleTheme = () => {
    if (isDark) {
      document.documentElement.classList.remove('dark');
      setIsDark(false);
    } else {
      document.documentElement.classList.add('dark');
      setIsDark(true);
    }
  };

  // RBAC Access Control Guard
  if (!isAuthenticated || !isAdmin) {
    return (
      <div className="min-h-screen flex items-center justify-center p-6 bg-background text-foreground">
        <div className="max-w-md w-full p-8 border border-border bg-card rounded-2xl shadow-elevated text-center space-y-5">
          <div className="w-16 h-16 rounded-2xl bg-amber-500/10 text-amber-600 flex items-center justify-center mx-auto border border-amber-500/20">
            <ShieldX className="w-8 h-8" />
          </div>
          <div>
            <h1 className="text-2xl font-bold font-editorial tracking-tight text-foreground">
              Administrator Privileges Required
            </h1>
            <p className="text-xs text-muted-foreground mt-2 leading-relaxed">
              Administrative credentials are required to access management controls. Please sign in with an authorized administrator account.
            </p>
          </div>

          <div className="p-3.5 rounded-xl bg-surface border border-border/80 text-left space-y-1">
            <p className="text-xs font-semibold text-foreground flex items-center gap-1.5">
              <Sparkles className="w-3.5 h-3.5 text-primary" />
              <span>Admin Access:</span>
            </p>
            <p className="text-[11px] text-muted-foreground leading-relaxed">
              Sign in with <span className="font-mono font-semibold text-foreground">admin@stratum.com</span> to access operational management controls.
            </p>
          </div>

          <div className="pt-2 flex flex-col gap-2">
            <Button
              onClick={() => navigate('/login?redirect=/admin')}
              variant="primary"
              size="md"
              leftIcon={<ShieldCheck className="w-4 h-4" />}
            >
              Sign In as Administrator
            </Button>
            <div className="grid grid-cols-2 gap-2">
              <Button onClick={() => navigate('/register?redirect=/admin')} variant="outline" size="sm">
                Create Admin Account
              </Button>
              <Button onClick={() => navigate('/')} variant="ghost" size="sm">
                Return to Store
              </Button>
            </div>
          </div>
        </div>
      </div>
    );
  }

  const navItems = [
    { label: 'Overview', path: '/admin', icon: LayoutDashboard },
    { label: 'Customer Orders', path: '/admin/orders', icon: ShoppingCart },
    { label: 'Inventory Management', path: '/admin/inventory', icon: Package },
    { label: 'Operations Analytics', path: '/admin/analytics', icon: BarChart3 },
  ];

  return (
    <div className="min-h-screen flex bg-background text-foreground transition-colors duration-200">
      {/* 1. Desktop Enterprise Sidebar */}
      <aside
        className={`fixed inset-y-0 left-0 z-50 w-64 bg-card border-r border-border flex flex-col transition-transform duration-200 md:translate-x-0 ${
          isSidebarOpen ? 'translate-x-0' : '-translate-x-full'
        }`}
      >
        {/* Brand Header */}
        <div className="h-16 flex items-center justify-between px-6 border-b border-border">
          <Link to="/admin" className="flex items-center gap-2.5">
            <div className="w-8 h-8 rounded-lg bg-primary text-primary-foreground flex items-center justify-center font-bold text-sm tracking-wider shadow-sm">
              S
            </div>
            <div>
              <span className="font-bold text-base tracking-tight text-foreground block">
                STRATUM
              </span>
              <span className="text-[10px] text-muted-foreground tracking-widest uppercase font-mono block -mt-0.5">
                Operations
              </span>
            </div>
          </Link>
          <button
            onClick={() => setIsSidebarOpen(false)}
            className="md:hidden p-1 text-muted-foreground hover:text-foreground"
            aria-label="Close sidebar"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Navigation Items */}
        <nav className="flex-1 p-4 space-y-1.5 overflow-y-auto">
          {navItems.map((item) => {
            const isActive = location.pathname === item.path;
            const Icon = item.icon;
            return (
              <Link
                key={item.path}
                to={item.path}
                onClick={() => setIsSidebarOpen(false)}
                className={`flex items-center gap-3 px-3.5 py-2.5 rounded-lg text-xs font-semibold uppercase tracking-wider transition-all ${
                  isActive
                    ? 'bg-primary text-primary-foreground shadow-sm'
                    : 'text-muted-foreground hover:bg-surface hover:text-foreground'
                }`}
              >
                <Icon className="w-4 h-4 shrink-0" />
                <span>{item.label}</span>
              </Link>
            );
          })}
        </nav>

        {/* Sidebar Footer */}
        <div className="p-4 border-t border-border space-y-3">
          <Link
            to="/"
            className="flex items-center gap-2 text-xs font-medium text-muted-foreground hover:text-foreground p-2 rounded-lg hover:bg-surface transition-colors"
          >
            <ArrowLeft className="w-4 h-4" />
            <span>Customer Storefront</span>
          </Link>

          <div className="flex items-center justify-between pt-2 border-t border-border/60">
            <div className="truncate pr-2">
              <p className="text-xs font-semibold text-foreground truncate">{customerID}</p>
              <span className="inline-flex items-center gap-1 text-[10px] text-emerald-600 dark:text-emerald-400 font-mono">
                <ShieldCheck className="w-3 h-3" /> Admin Verified
              </span>
            </div>
            <button
              onClick={logout}
              className="p-1.5 text-muted-foreground hover:text-red-500 rounded-md hover:bg-surface transition-colors"
              title="Sign out"
              aria-label="Sign out"
            >
              <LogOut className="w-4 h-4" />
            </button>
          </div>
        </div>
      </aside>

      {/* 2. Main Content Area */}
      <div className="flex-1 flex flex-col md:pl-64 min-w-0">
        {/* Top Operational Bar */}
        <header className="sticky top-0 z-30 h-16 bg-card/90 backdrop-blur-md border-b border-border flex items-center justify-between px-4 sm:px-6">
          <div className="flex items-center gap-3">
            <button
              onClick={() => setIsSidebarOpen(true)}
              className="md:hidden p-2 text-muted-foreground hover:text-foreground rounded-lg hover:bg-surface"
              aria-label="Open sidebar"
              aria-expanded={isSidebarOpen}
            >
              <Menu className="w-5 h-5" />
            </button>
            <div className="flex items-center gap-2 text-xs text-muted-foreground">
              <span>Admin Console</span>
              <span>/</span>
              <span className="font-semibold text-foreground capitalize">
                {location.pathname.replace('/admin', '').replace('/', '') || 'Overview'}
              </span>
            </div>
          </div>

          <div className="flex items-center gap-3">
            <span className="hidden sm:inline-flex items-center gap-1.5 text-[11px] font-medium px-2.5 py-1 rounded-full bg-emerald-500/10 text-emerald-700 dark:text-emerald-400 border border-emerald-500/20">
              <span className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse" />
              Operational · Live
            </span>
            <button
              onClick={toggleTheme}
              className="p-2 rounded-lg text-muted-foreground hover:text-foreground hover:bg-surface transition-colors"
              aria-label="Toggle theme"
            >
              {isDark ? <Sun className="w-4 h-4" /> : <Moon className="w-4 h-4" />}
            </button>
          </div>
        </header>

        {/* Admin Route Outlet */}
        <main className="flex-1 p-4 sm:p-6 lg:p-8 max-w-7xl w-full mx-auto">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
