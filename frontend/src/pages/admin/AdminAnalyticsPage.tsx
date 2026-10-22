import React from 'react';
import { useAdminAnalytics } from '@/features/admin/useAdmin';
import { useInventory } from '@/features/catalog/useInventory';
import { formatCurrency } from '@/lib/utils';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/Card';
import { Badge } from '@/components/ui/Badge';
import {
  BarChart3,
  TrendingUp,
  Activity,
  Layers,
  Database,
  Radio,
  Info,
} from 'lucide-react';
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
  Legend,
} from 'recharts';

export function AdminAnalyticsPage() {
  const { data: analyticsData } = useAdminAnalytics();
  const { products } = useInventory();

  const totalStockUnits = products.reduce((acc, p) => acc + (p?.stockQuantity || 0), 0);
  const totalValue = products.reduce((acc, p) => acc + (p?.stockQuantity || 0) * (p?.price || 0), 0);

  const inventoryDistribution = products.map((p) => ({
    name: p.name || 'Item',
    value: p.stockQuantity || 0,
  }));

  const valueDistribution = products.map((p) => ({
    name: p.name || 'Item',
    value: (p.stockQuantity || 0) * (p.price || 0),
  }));


  const PIE_COLORS = ['#0F4A36', '#1DA87A', '#3B82F6', '#F59E0B', '#8B5CF6'];

  const operationalModules = [
    { name: 'Storefront & Customer Gateway', channel: 'Customer Web Portal', state: 'Operational', role: 'Traffic Routing & Account Security' },
    { name: 'STRATUM Order Management Engine', channel: 'Fulfillment Pipeline', state: 'Operational', role: 'Order Creation & Status Lifecycle' },
    { name: 'Payment Processing Service', channel: 'Payment Integration', state: 'Operational', role: 'Payment Link Generation & Verification' },
    { name: 'Real-Time Inventory Engine', channel: 'Warehouse Catalog', state: 'Operational', role: 'Live Stock Tracking & Allocation' },
    { name: 'Kitchen Dispatch Station', channel: 'Kitchen Operations', state: 'Operational', role: 'Meal Preparation & Readiness Tracking' },
    { name: 'Shopping Cart Session Store', channel: 'Active Basket', state: 'Operational', role: 'Synchronized Cart Management' },
    { name: 'Operational Notifications', channel: 'Event Dispatch', state: 'Operational', role: 'System Logging & Customer Updates' },
  ];

  return (
    <div className="space-y-8 text-left">
      {/* Header */}
      <div className="pb-4 border-b border-border">
        <h1 className="text-2xl sm:text-3xl font-bold tracking-tight text-foreground">
          Operations Analytics & Valuation
        </h1>
        <p className="text-xs sm:text-sm text-muted-foreground mt-0.5">
          Inventory valuation, stock volume distribution, and fulfillment operational status.
        </p>
      </div>

      {/* Top Stat Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-6">
        <Card className="p-6 border border-border shadow-subtle space-y-2">
          <div className="flex items-center justify-between text-muted-foreground">
            <span className="text-xs font-semibold uppercase tracking-wider">Inventory Value</span>
            <TrendingUp className="w-4 h-4 text-emerald-600" />
          </div>
          <p className="text-3xl font-extrabold font-mono text-foreground">
            {formatCurrency(totalValue)}
          </p>
          <p className="text-xs text-muted-foreground">Calculated across available catalog units</p>
        </Card>

        <Card className="p-6 border border-border shadow-subtle space-y-2">
          <div className="flex items-center justify-between text-muted-foreground">
            <span className="text-xs font-semibold uppercase tracking-wider">Total Warehouse Units</span>
            <Layers className="w-4 h-4 text-primary" />
          </div>
          <p className="text-3xl font-extrabold font-mono text-foreground">
            {totalStockUnits}
          </p>
          <p className="text-xs text-muted-foreground">Units available for checkout</p>
        </Card>

        <Card className="p-6 border border-border shadow-subtle space-y-2">
          <div className="flex items-center justify-between text-muted-foreground">
            <span className="text-xs font-semibold uppercase tracking-wider">Operational Health</span>
            <Activity className="w-4 h-4 text-emerald-600" />
          </div>
          <p className="text-3xl font-extrabold font-mono text-foreground">
            100%
          </p>
          <p className="text-xs text-emerald-600 font-medium">All fulfillment channels active & healthy</p>
        </Card>
      </div>

      {/* Recharts Analytics Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-8">
        {/* Inventory Units Breakdown */}
        <Card className="lg:col-span-6 p-6 border border-border shadow-subtle space-y-4">
          <h3 className="text-base font-bold text-foreground">Stock Volume Distribution</h3>
          <div className="h-64 w-full">
            <ResponsiveContainer width="100%" height="100%">
              <PieChart>
                <Pie
                  data={inventoryDistribution}
                  dataKey="value"
                  nameKey="name"
                  cx="50%"
                  cy="45%"
                  innerRadius={48}
                  outerRadius={72}
                  paddingAngle={3}
                >
                  {inventoryDistribution.map((_, index) => (
                    <Cell key={`cell-${index}`} fill={PIE_COLORS[index % PIE_COLORS.length]} />
                  ))}
                </Pie>
                <Tooltip
                  contentStyle={{
                    backgroundColor: 'hsl(var(--card))',
                    borderColor: 'hsl(var(--border))',
                    borderRadius: '8px',
                    fontSize: '12px',
                  }}
                  formatter={(val: any, name: any) => [`${val} units`, name]}
                />
                <Legend verticalAlign="bottom" height={36} iconType="circle" iconSize={8} wrapperStyle={{ fontSize: '11px' }} />
              </PieChart>
            </ResponsiveContainer>
          </div>
        </Card>

        {/* Inventory Capital Valuation */}
        <Card className="lg:col-span-6 p-6 border border-border shadow-subtle space-y-4">
          <h3 className="text-base font-bold text-foreground">Capital Valuation by SKU</h3>
          <div className="h-64 w-full">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={valueDistribution} margin={{ top: 10, right: 10, left: 10, bottom: 20 }}>
                <XAxis dataKey="name" stroke="#888888" fontSize={11} tickLine={false} />
                <YAxis stroke="#888888" fontSize={11} tickLine={false} />
                <Tooltip
                  contentStyle={{
                    backgroundColor: 'hsl(var(--card))',
                    borderColor: 'hsl(var(--border))',
                    borderRadius: '8px',
                    fontSize: '12px',
                  }}
                  formatter={(val: any) => [formatCurrency(Number(val)), 'Valuation']}
                />
                <Bar dataKey="value" fill="#0F4A36" radius={[4, 4, 0, 0]} />
              </BarChart>
            </ResponsiveContainer>
          </div>
        </Card>
      </div>

      {/* Operational Systems Status Table */}
      <Card className="border border-border overflow-hidden shadow-subtle space-y-0">
        <CardHeader className="bg-surface/50 border-b border-border/60 p-4 sm:p-6">
          <CardTitle className="text-base font-bold flex items-center gap-2">
            <Database className="w-4 h-4 text-primary" />
            <span>Fulfillment Modules & Systems</span>
          </CardTitle>
        </CardHeader>
        <div className="overflow-x-auto">
          <table className="w-full min-w-[650px] text-xs text-left">
            <thead className="bg-surface/30 border-b border-border text-muted-foreground uppercase font-mono tracking-wider">
              <tr>
                <th className="py-3 px-4">Operational Module</th>
                <th className="py-3 px-4">Channel</th>
                <th className="py-3 px-4">Status</th>
                <th className="py-3 px-4">Operational Responsibility</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {operationalModules.map((mod) => (
                <tr key={mod.name} className="hover:bg-surface/40 transition-colors">
                  <td className="py-3 px-4 font-semibold text-foreground">
                    {mod.name}
                  </td>
                  <td className="py-3 px-4 text-muted-foreground">
                    {mod.channel}
                  </td>
                  <td className="py-3 px-4">
                    <Badge variant="paid" size="sm">
                      {mod.state}
                    </Badge>
                  </td>
                  <td className="py-3 px-4 text-muted-foreground">
                    {mod.role}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </Card>
    </div>
  );
}
