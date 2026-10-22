import React from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useAuthStore } from '@/features/auth/useAuthStore';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { User, Package, ShieldAlert, LogOut, ArrowRight } from 'lucide-react';

export function AccountPage() {
  const navigate = useNavigate();
  const { customerID, role, isAdmin, logout } = useAuthStore();

  return (
    <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-8 text-left">
      <div>
        <h1 className="text-3xl font-bold font-editorial tracking-tight text-foreground">
          My Account
        </h1>
        <p className="text-sm text-muted-foreground mt-1">
          Manage your customer profile and platform preferences.
        </p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-12 gap-8">
        <div className="md:col-span-6 space-y-6">
          <Card className="p-6 border border-border shadow-subtle space-y-4">
            <div className="flex items-center gap-4">
              <div className="w-14 h-14 rounded-full bg-primary-light text-primary flex items-center justify-center font-bold text-xl">
                {customerID ? customerID[0].toUpperCase() : 'U'}
              </div>
              <div>
                <p className="font-bold text-base text-foreground truncate max-w-[220px]">
                  {customerID}
                </p>
                <div className="mt-1">
                  <Badge variant={isAdmin ? 'primary' : 'neutral'} size="sm">
                    {isAdmin ? 'System Administrator' : 'Verified Customer'}
                  </Badge>
                </div>
              </div>
            </div>

            <div className="pt-4 border-t border-border space-y-2 text-xs text-muted-foreground">
              <p>Account Status: <span className="font-medium text-emerald-600 dark:text-emerald-400">Active & Verified</span></p>
              <p>Security: <span className="font-medium text-foreground">Encrypted Session</span></p>
            </div>

            <div className="pt-2">
              <Button onClick={logout} variant="outline" size="sm" className="w-full text-red-500 hover:text-red-600" leftIcon={<LogOut className="w-4 h-4" />}>
                Sign Out
              </Button>
            </div>
          </Card>
        </div>

        <div className="md:col-span-6 space-y-4">
          <Card className="p-6 border border-border shadow-subtle space-y-3">
            <div className="w-10 h-10 rounded-lg bg-emerald-500/10 text-emerald-600 flex items-center justify-center">
              <Package className="w-5 h-5" />
            </div>
            <h3 className="font-bold text-base text-foreground">Order Tracking</h3>
            <p className="text-xs text-muted-foreground leading-relaxed">
              Check real-time order states, tracking details, and fulfillment lifecycle history.
            </p>
            <Button onClick={() => navigate('/orders')} variant="primary" size="sm" rightIcon={<ArrowRight className="w-3.5 h-3.5" />}>
              View Order History
            </Button>
          </Card>

          {isAdmin && (
            <Card className="p-6 border border-amber-500/20 bg-amber-500/5 shadow-subtle space-y-3">
              <div className="w-10 h-10 rounded-lg bg-amber-500/10 text-amber-700 dark:text-amber-400 flex items-center justify-center">
                <ShieldAlert className="w-5 h-5" />
              </div>
              <h3 className="font-bold text-base text-foreground">Operations Console</h3>
              <p className="text-xs text-muted-foreground leading-relaxed">
                You have administrative access to manage inventory, inspect orders, and review analytics.
              </p>
              <Button onClick={() => navigate('/admin')} variant="secondary" size="sm" rightIcon={<ArrowRight className="w-3.5 h-3.5" />}>
                Go to Operations Dashboard
              </Button>
            </Card>
          )}
        </div>
      </div>
    </div>
  );
}
