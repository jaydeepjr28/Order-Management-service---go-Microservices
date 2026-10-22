import React, { useState } from 'react';
import { Outlet, Link, useLocation, useNavigate } from 'react-router-dom';
import { useAuthStore } from '@/features/auth/useAuthStore';
import { useCart } from '@/features/cart/useCart';
import { SITE_CONFIG } from '@/config/site';
import { Button } from '@/components/ui/Button';
import {
  ShoppingBag,
  User,
  LogOut,
  ShieldAlert,
  Menu,
  X,
  Sun,
  Moon,
  ChevronRight,
} from 'lucide-react';

export function StorefrontLayout() {
  const location = useLocation();
  const navigate = useNavigate();
  const { isAuthenticated, isAdmin, customerID, logout } = useAuthStore();
  const { totalQuantity } = useCart();
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
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

  const navLinks = [
    { label: 'Home', path: '/' },
    { label: 'Catalog', path: '/products' },
    ...(isAuthenticated && !isAdmin ? [{ label: 'My Orders', path: '/orders' }] : []),
  ];

  return (
    <div className="min-h-screen flex flex-col bg-background text-foreground transition-colors duration-200">
      {/* 1. Announcement Bar */}
      <div className="bg-primary text-primary-foreground text-xs py-2 px-4 text-center font-medium tracking-wide">
        <span>{SITE_CONFIG.announcement}</span>
      </div>

      {/* 2. Responsive Navigation Bar */}
      <header className="sticky top-0 z-40 bg-card/90 backdrop-blur-md border-b border-border shadow-subtle transition-all">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
          {/* Brand Logo & Name */}
          <div className="flex items-center gap-8">
            <Link to="/" className="flex items-center gap-2.5 group">
              <div className="w-9 h-9 rounded-lg bg-primary text-primary-foreground flex items-center justify-center font-bold text-lg tracking-wider shadow-sm group-hover:scale-105 transition-transform">
                S
              </div>
              <span className="font-extrabold text-xl tracking-tight text-foreground">
                {SITE_CONFIG.name}
              </span>
            </Link>

            {/* Desktop Navigation */}
            <nav className="hidden md:flex items-center gap-6">
              {navLinks.map((link) => {
                const isActive = location.pathname === link.path;
                return (
                  <Link
                    key={link.path}
                    to={link.path}
                    className={`text-sm font-medium transition-colors ${
                      isActive
                        ? 'text-primary font-semibold'
                        : 'text-muted-foreground hover:text-foreground'
                    }`}
                  >
                    {link.label}
                  </Link>
                );
              })}
              {isAdmin && (
                <Link
                  to="/admin"
                  className="inline-flex items-center gap-1.5 text-xs font-semibold px-2.5 py-1 rounded-md bg-amber-500/10 text-amber-700 dark:text-amber-400 border border-amber-500/20 hover:bg-amber-500/20 transition-colors"
                >
                  <ShieldAlert className="w-3.5 h-3.5" />
                  Admin Console
                </Link>
              )}
            </nav>
          </div>

          {/* Right Header Actions */}
          <div className="flex items-center gap-3">
            {/* Theme Toggle */}
            <button
              onClick={toggleTheme}
              className="p-2 rounded-lg text-muted-foreground hover:text-foreground hover:bg-surface transition-colors"
              aria-label="Toggle theme"
            >
              {isDark ? <Sun className="w-4 h-4" /> : <Moon className="w-4 h-4" />}
            </button>

            {/* Cart Button (Customer Only) */}
            {!isAdmin && (
              <Link
                to="/cart"
                className="relative p-2 rounded-lg text-muted-foreground hover:text-foreground hover:bg-surface transition-colors flex items-center justify-center"
                aria-label="View shopping cart"
              >
                <ShoppingBag className="w-5 h-5" />
                {totalQuantity > 0 && (
                  <span className="absolute -top-1 -right-1 bg-primary text-primary-foreground text-[11px] font-bold h-5 min-w-[20px] px-1 rounded-full flex items-center justify-center shadow-sm">
                    {totalQuantity}
                  </span>
                )}
              </Link>
            )}

            {/* Auth Actions (Desktop) */}
            <div className="hidden sm:flex items-center gap-2 pl-2 border-l border-border">
              {isAuthenticated ? (
                <div className="flex items-center gap-2">
                  <div className="text-right">
                    <p className="text-xs font-semibold text-foreground truncate max-w-[130px]">
                      {customerID}
                    </p>
                    <p className="text-[10px] text-muted-foreground uppercase font-mono">
                      {isAdmin ? 'Admin' : 'Customer'}
                    </p>
                  </div>
                  <Button
                    onClick={logout}
                    variant="ghost"
                    size="sm"
                    className="p-2 text-muted-foreground hover:text-red-500"
                    title="Sign out"
                    aria-label="Sign out"
                  >
                    <LogOut className="w-4 h-4" />
                  </Button>
                </div>
              ) : (
                <div className="flex items-center gap-2">
                  <Button
                    onClick={() => navigate('/login')}
                    variant="ghost"
                    size="sm"
                  >
                    Sign in
                  </Button>
                  <Button
                    onClick={() => navigate('/register')}
                    variant="primary"
                    size="sm"
                  >
                    Register
                  </Button>
                </div>
              )}
            </div>

            {/* Mobile Menu Toggle */}
            <button
              onClick={() => setIsMobileMenuOpen(!isMobileMenuOpen)}
              className="p-2 rounded-lg md:hidden text-muted-foreground hover:text-foreground hover:bg-surface transition-colors"
              aria-label="Toggle mobile menu"
              aria-expanded={isMobileMenuOpen}
            >
              {isMobileMenuOpen ? <X className="w-5 h-5" /> : <Menu className="w-5 h-5" />}
            </button>
          </div>
        </div>

        {/* Mobile Navigation Drawer */}
        {isMobileMenuOpen && (
          <div className="md:hidden border-b border-border bg-card px-4 pt-2 pb-6 space-y-3 animate-in slide-in-from-top-2 duration-200">
            <nav className="flex flex-col space-y-1">
              {navLinks.map((link) => (
                <Link
                  key={link.path}
                  to={link.path}
                  onClick={() => setIsMobileMenuOpen(false)}
                  className="px-3 py-2 rounded-lg text-sm font-medium text-foreground hover:bg-surface flex items-center justify-between"
                >
                  <span>{link.label}</span>
                  <ChevronRight className="w-4 h-4 text-muted-foreground" />
                </Link>
              ))}
              {isAdmin && (
                <Link
                  to="/admin"
                  onClick={() => setIsMobileMenuOpen(false)}
                  className="px-3 py-2 rounded-lg text-sm font-semibold text-amber-700 dark:text-amber-400 bg-amber-500/10 flex items-center justify-between"
                >
                  <span className="flex items-center gap-2">
                    <ShieldAlert className="w-4 h-4" /> Admin Operations
                  </span>
                  <ChevronRight className="w-4 h-4" />
                </Link>
              )}
            </nav>

            <div className="pt-3 border-t border-border">
              {isAuthenticated ? (
                <div className="flex items-center justify-between px-3">
                  <div>
                    <p className="text-xs font-semibold text-foreground">{customerID}</p>
                    <p className="text-[11px] text-muted-foreground">{isAdmin ? 'Administrator' : 'Customer'}</p>
                  </div>
                  <Button onClick={() => { logout(); setIsMobileMenuOpen(false); }} variant="outline" size="sm">
                    Sign Out
                  </Button>
                </div>
              ) : (
                <div className="grid grid-cols-2 gap-2">
                  <Button onClick={() => { navigate('/login'); setIsMobileMenuOpen(false); }} variant="outline" size="sm">
                    Sign In
                  </Button>
                  <Button onClick={() => { navigate('/register'); setIsMobileMenuOpen(false); }} variant="primary" size="sm">
                    Register
                  </Button>
                </div>
              )}
            </div>
          </div>
        )}
      </header>

      {/* 3. Main Outlet Container */}
      <main className="flex-1 flex flex-col">
        <Outlet />
      </main>

      {/* 4. Refined Storefront Footer */}
      <footer className="border-t border-border bg-card/50 py-12 mt-16 text-sm text-muted-foreground">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 grid grid-cols-1 md:grid-cols-4 gap-8">
          <div className="space-y-3 md:col-span-2">
            <div className="flex items-center gap-2 font-bold text-lg text-foreground">
              <span className="w-6 h-6 rounded bg-primary text-primary-foreground flex items-center justify-center text-xs">S</span>
              {SITE_CONFIG.fullName}
            </div>
            <p className="text-xs leading-relaxed max-w-sm">
              Artisanal culinary provisions crafted daily with exceptional ingredients, uncompromising quality, and fast, seamless fulfillment.
            </p>
          </div>
          <div className="space-y-2">
            <h4 className="text-xs font-semibold uppercase tracking-wider text-foreground">Quick Links</h4>
            <ul className="space-y-1.5 text-xs">
              <li><Link to="/" className="hover:text-foreground transition-colors">Home</Link></li>
              <li><Link to="/products" className="hover:text-foreground transition-colors">Catalog</Link></li>
              <li><Link to="/cart" className="hover:text-foreground transition-colors">Cart</Link></li>
              <li><Link to="/orders" className="hover:text-foreground transition-colors">Order Tracking</Link></li>
            </ul>
          </div>
          <div className="space-y-2">
            <h4 className="text-xs font-semibold uppercase tracking-wider text-foreground">Quality Guarantee</h4>
            <p className="text-xs">Handcrafted Fresh Daily</p>
            <p className="text-xs">Sustainably Sourced Ingredients</p>
            <p className="text-xs">Contactless Fast Fulfillment</p>
          </div>
        </div>
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pt-8 mt-8 border-t border-border flex flex-col sm:flex-row items-center justify-between text-xs">
          <p>© {new Date().getFullYear()} {SITE_CONFIG.fullName}. All rights reserved.</p>
          <p className="mt-2 sm:mt-0 font-medium text-[11px] text-emerald-600 dark:text-emerald-400">All Systems Operational</p>
        </div>
      </footer>
    </div>
  );
}
