import React from 'react';
import { useNavigate } from 'react-router-dom';
import { useAdminAnalytics } from '@/features/admin/useAdmin';
import { useInventory } from '@/features/catalog/useInventory';
import { formatCurrency } from '@/lib/utils';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { Skeleton } from '@/components/ui/Skeleton';
import {
  Package,
  Layers,
  Activity,
  ArrowRight,
  Info,
  Server,
  Zap,
} from 'lucide-react';
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  Cell,
} from 'recharts';

export function AdminOverviewPage() {
  const navigate = useNavigate();
  const { data: analyticsData, isLoading: isAnalyticsLoading } = useAdminAnalytics();
  const { products, isLoading: isInventoryLoading } = useInventory();

  const totalStockUnits = products.reduce((acc, p) => acc + p.stockQuantity, 0);
  const totalCatalogItems = products.length;

  // Chart data from actual live inventory
  const chartData = products.map((p) => ({
    name: p.name,
    stock: p.stockQuantity,
    price: p.price,
  }));

  const COLORS = ['#0F4A36', '#1DA87A', '#3B82F6', '#F59E0B'];

  return (
    <div className="space-y-8 text-left">
      {/* Page Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between pb-4 border-b border-border gap-4">
        <div>
          <h1 className="text-2xl sm:text-3xl font-bold tracking-tight text-foreground">
            Operations Overview
          </h1>
          <p className="text-xs sm:text-sm text-muted-foreground mt-0.5">
            Real-time operations, inventory monitoring, and fulfillment overview.
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button
            onClick={() => navigate('/admin/inventory')}
            variant="primary"
            size="sm"
          >
            Manage Inventory
          </Button>
          <Button
            onClick={() => navigate('/admin/orders')}
            variant="outline"
            size="sm"
          >
            View Orders
          </Button>
        </div>
      </div>

      {/* Metric Cards Grid */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 sm:gap-6">
        <Card className="p-5 border border-border shadow-subtle space-y-2">
          <div className="flex items-center justify-between text-muted-foreground">
            <span className="text-xs font-semibold uppercase tracking-wider">Catalog SKUs</span>
            <Package className="w-4 h-4 text-primary" />
          </div>
          {isInventoryLoading ? (
            <Skeleton className="h-8 w-16" />
          ) : (
            <p className="text-2xl sm:text-3xl font-extrabold font-mono text-foreground">
              {totalCatalogItems}
            </p>
          )}
          <p className="text-[11px] text-muted-foreground">Active in store catalog</p>
        </Card>

        <Card className="p-5 border border-border shadow-subtle space-y-2">
          <div className="flex items-center justify-between text-muted-foreground">
            <span className="text-xs font-semibold uppercase tracking-wider">Available Units</span>
            <Layers className="w-4 h-4 text-emerald-600" />
          </div>
          {isInventoryLoading ? (
            <Skeleton className="h-8 w-20" />
          ) : (
            <p className="text-2xl sm:text-3xl font-extrabold font-mono text-foreground">
              {totalStockUnits}
            </p>
          )}
          <p className="text-[11px] text-muted-foreground">Aggregate units in warehouse</p>
        </Card>

        <Card className="p-5 border border-border shadow-subtle space-y-2">
          <div className="flex items-center justify-between text-muted-foreground">
            <span className="text-xs font-semibold uppercase tracking-wider">System Status</span>
            <Server className="w-4 h-4 text-emerald-600" />
          </div>
          <p className="text-2xl sm:text-3xl font-extrabold font-mono text-foreground">
            100%
          </p>
          <p className="text-[11px] text-emerald-600 font-medium">All Systems Operational</p>
        </Card>

        <Card className="p-5 border border-border shadow-subtle space-y-2">
          <div className="flex items-center justify-between text-muted-foreground">
            <span className="text-xs font-semibold uppercase tracking-wider">Fulfillment</span>
            <Zap className="w-4 h-4 text-amber-500" />
          </div>
          <p className="text-2xl sm:text-3xl font-extrabold font-mono text-foreground">
            Real-Time
          </p>
          <p className="text-[11px] text-muted-foreground">Automated Order Pipeline</p>
        </Card>
      </div>

      {/* Main Charts & Breakdown */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-8">
        {/* Stock Distribution Chart */}
        <Card className="lg:col-span-8 p-6 border border-border shadow-subtle space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <h3 className="text-base font-bold text-foreground">Live Stock Levels by SKU</h3>
              <p className="text-xs text-muted-foreground">
                Live stock levels across all catalog items
              </p>
            </div>
            <Badge variant="primary" size="sm">
              Live Feed
            </Badge>
          </div>

          <div className="h-64 w-full pt-4">
            {isInventoryLoading ? (
              <Skeleton className="h-full w-full rounded-xl" />
            ) : chartData.length > 0 ? (
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={chartData} margin={{ top: 10, right: 10, left: -20, bottom: 20 }}>
                  <XAxis dataKey="name" stroke="#888888" fontSize={11} tickLine={false} />
                  <YAxis stroke="#888888" fontSize={11} tickLine={false} />
                  <Tooltip
                    contentStyle={{
                      backgroundColor: 'hsl(var(--card))',
                      borderColor: 'hsl(var(--border))',
                      borderRadius: '8px',
                      fontSize: '12px',
                    }}
                  />
                  <Bar dataKey="stock" radius={[4, 4, 0, 0]}>
                    {chartData.map((_, index) => (
                      <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                    ))}
                  </Bar>
                </BarChart>
              </ResponsiveContainer>
            ) : (
              <div className="h-full flex items-center justify-center text-xs text-muted-foreground">
                No inventory data reported.
              </div>
            )}
          </div>
        </Card>

        {/* Quick Operations Actions */}
        <Card className="lg:col-span-4 p-6 border border-border shadow-subtle space-y-4 flex flex-col justify-between">
          <div className="space-y-4">
            <h3 className="text-base font-bold text-foreground">Operational Actions</h3>
            <p className="text-xs text-muted-foreground leading-relaxed">
              Manage inventory, track customer orders, and review store performance.
            </p>

            <div className="space-y-2.5">
              <button
                onClick={() => navigate('/admin/inventory')}
                className="w-full p-3 rounded-lg border border-border bg-surface hover:bg-surface-hover text-left flex items-center justify-between group transition-colors"
              >
                <div>
                  <p className="text-xs font-bold text-foreground">Adjust Stock Quantities</p>
                  <p className="text-[11px] text-muted-foreground">Manage active warehouse items</p>
                </div>
                <ArrowRight className="w-4 h-4 text-muted-foreground group-hover:translate-x-0.5 transition-transform" />
              </button>

              <button
                onClick={() => navigate('/admin/orders')}
                className="w-full p-3 rounded-lg border border-border bg-surface hover:bg-surface-hover text-left flex items-center justify-between group transition-colors"
              >
                <div>
                  <p className="text-xs font-bold text-foreground">Inspect Orders</p>
                  <p className="text-[11px] text-muted-foreground">View customer order history and statuses</p>
                </div>
                <ArrowRight className="w-4 h-4 text-muted-foreground group-hover:translate-x-0.5 transition-transform" />
              </button>

              <button
                onClick={() => navigate('/admin/analytics')}
                className="w-full p-3 rounded-lg border border-border bg-surface hover:bg-surface-hover text-left flex items-center justify-between group transition-colors"
              >
                <div>
                  <p className="text-xs font-bold text-foreground">Operations Analytics</p>
                  <p className="text-[11px] text-muted-foreground">Fulfillment throughput and performance</p>
                </div>
                <ArrowRight className="w-4 h-4 text-muted-foreground group-hover:translate-x-0.5 transition-transform" />
              </button>
            </div>
          </div>

          <div className="pt-4 border-t border-border text-[11px] text-muted-foreground">
            Authenticated Administrator Session
          </div>
        </Card>
      </div>
    </div>
  );
}
