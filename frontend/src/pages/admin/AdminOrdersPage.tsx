import React, { useState } from 'react';
import { useAdminOrders, useAdminCancelOrder } from '@/features/admin/useAdmin';
import { useCustomerOrders } from '@/features/orders/useOrders';
import { Badge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import { Card } from '@/components/ui/Card';
import { Dialog } from '@/components/ui/Dialog';
import { Input } from '@/components/ui/Input';
import { OrderProgress } from '@/components/ui/OrderProgress';
import { formatCurrency } from '@/lib/utils';
import { getProductMeta } from '@/config/site';
import { Search, Eye, Info, RefreshCw, ShoppingCart, XCircle } from 'lucide-react';
import { Order } from '@/types';

export function AdminOrdersPage() {
  const { data: adminOrdersData, isLoading: isAdminLoading, refetch: refetchAdmin } = useAdminOrders();
  const { data: customerOrders, isLoading: isCustLoading } = useCustomerOrders();
  const adminCancelMutation = useAdminCancelOrder();

  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [selectedOrder, setSelectedOrder] = useState<Order | null>(null);


  // If backend returned actual array of orders or stub message
  const isStubMessage = adminOrdersData && !Array.isArray(adminOrdersData);

  // Use orders array from backend or active session orders
  const ordersList: Order[] = Array.isArray(adminOrdersData)
    ? adminOrdersData
    : (customerOrders || []);

  const filteredOrders = ordersList.filter((order: any) => {
    const id = String(order.ID || order.id || '');
    const custId = String(order.customerID || order.CustomerID || order.customerId || '');
    const st = String(order.Status || order.status || '');
    const matchesSearch =
      id.toLowerCase().includes(search.toLowerCase()) ||
      custId.toLowerCase().includes(search.toLowerCase());
    const matchesStatus =
      statusFilter === 'all' ||
      st.toLowerCase() === statusFilter.toLowerCase();
    return matchesSearch && matchesStatus;
  });

  return (
    <div className="space-y-6 text-left">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between pb-4 border-b border-border gap-4">
        <div>
          <h1 className="text-2xl sm:text-3xl font-bold tracking-tight text-foreground">
            Order Operations
          </h1>
          <p className="text-xs sm:text-sm text-muted-foreground mt-0.5">
            Manage, track, and inspect customer orders and fulfillment statuses.
          </p>
        </div>
        <Button
          onClick={() => refetchAdmin()}
          variant="outline"
          size="sm"
          leftIcon={<RefreshCw className="w-3.5 h-3.5" />}
        >
          Refresh Feed
        </Button>
      </div>

      {/* Filters Bar */}
      <div className="flex flex-col sm:flex-row gap-3 items-stretch sm:items-center justify-between">
        <div className="max-w-xs w-full">
          <Input
            placeholder="Search by Order ID or Customer..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            leftIcon={<Search className="w-4 h-4" />}
          />
        </div>

        <div className="flex items-center gap-1.5 overflow-x-auto pb-1 sm:pb-0">
          {['all', 'pending', 'paid', 'ready'].map((st) => (
            <button
              key={st}
              onClick={() => setStatusFilter(st)}
              className={`px-3 py-1.5 rounded-lg text-xs font-semibold capitalize transition-colors ${
                statusFilter === st
                  ? 'bg-primary text-primary-foreground'
                  : 'bg-surface text-muted-foreground hover:text-foreground'
              }`}
            >
              {st}
            </button>
          ))}
        </div>
      </div>

      {/* Orders Table */}
      <Card className="border border-border overflow-hidden shadow-subtle">
        <div className="overflow-x-auto">
          <table className="w-full min-w-[700px] text-xs text-left">
            <thead className="bg-surface/70 border-b border-border text-muted-foreground uppercase font-mono tracking-wider">
              <tr>
                <th className="py-3 px-4">Order ID</th>
                <th className="py-3 px-4">Customer Identity</th>
                <th className="py-3 px-4">Status</th>
                <th className="py-3 px-4">Items Summary</th>
                <th className="py-3 px-4">Payment Link</th>
                <th className="py-3 px-4 text-right">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {isAdminLoading || isCustLoading ? (
                Array.from({ length: 4 }).map((_, idx) => (
                  <tr key={idx} className="animate-pulse">
                    <td className="py-3.5 px-4"><div className="h-4 w-16 bg-muted rounded" /></td>
                    <td className="py-3.5 px-4"><div className="h-4 w-32 bg-muted rounded" /></td>
                    <td className="py-3.5 px-4"><div className="h-5 w-16 bg-muted rounded-full" /></td>
                    <td className="py-3.5 px-4"><div className="h-4 w-20 bg-muted rounded" /></td>
                    <td className="py-3.5 px-4"><div className="h-4 w-14 bg-muted rounded" /></td>
                    <td className="py-3.5 px-4 text-right"><div className="h-7 w-16 bg-muted rounded ml-auto" /></td>
                  </tr>
                ))
              ) : filteredOrders.length > 0 ? (
                filteredOrders.map((order: any) => {
                  const orderId = order.ID || order.id || '';
                  const customerEmail = order.customerID || order.CustomerID || order.customerId || 'Customer';
                  const orderStatus = order.Status || order.status || 'pending';
                  const status = orderStatus.toLowerCase();
                  const badgeVariant =
                    status === 'ready'
                      ? 'ready'
                      : status === 'paid'
                      ? 'paid'
                      : 'pending';

                  const items = order.Items || order.items || [];
                  const paymentLink = order.PaymentLink || order.paymentLink;

                  return (
                    <tr key={orderId} className="hover:bg-surface/40 transition-colors">
                      <td className="py-3.5 px-4 font-mono font-bold text-foreground">
                        #{orderId}
                      </td>
                      <td className="py-3.5 px-4 font-mono text-muted-foreground">
                        {customerEmail}
                      </td>
                      <td className="py-3.5 px-4">
                        <Badge variant={badgeVariant} size="sm">
                          {orderStatus}
                        </Badge>
                      </td>
                      <td className="py-3.5 px-4 text-muted-foreground">
                        {items.length} {items.length === 1 ? 'item' : 'items'}
                      </td>
                      <td className="py-3.5 px-4 font-mono">
                        {paymentLink ? (
                          <span className="text-emerald-600 font-semibold truncate max-w-[120px] block">
                            Generated
                          </span>
                        ) : (
                          <span className="text-muted-foreground italic">None</span>
                        )}
                      </td>
                      <td className="py-3.5 px-4 text-right">
                        <Button
                          onClick={() => setSelectedOrder(order)}
                          variant="ghost"
                          size="sm"
                          className="h-8 px-2.5 text-primary"
                          leftIcon={<Eye className="w-3.5 h-3.5" />}
                        >
                          Inspect
                        </Button>
                      </td>
                    </tr>
                  );
                })
              ) : (
                <tr>
                  <td colSpan={6} className="py-12 text-center text-muted-foreground">
                    <ShoppingCart className="w-6 h-6 mx-auto mb-2 opacity-40" />
                    No orders matching your criteria.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </Card>

      {/* Order Detail Modal */}
      {selectedOrder && (() => {
        const selOrder = selectedOrder as any;
        const selId = selOrder.ID || selOrder.id || '';
        const selCustomer = selOrder.customerID || selOrder.CustomerID || selOrder.customerId || '';
        const selStatus = selOrder.Status || selOrder.status || 'pending';
        const selItems = selOrder.Items || selOrder.items || [];
        const selPaymentLink = selOrder.PaymentLink || selOrder.paymentLink;

        return (
          <Dialog
            isOpen={!!selectedOrder}
            onClose={() => setSelectedOrder(null)}
            title={`Order #${selId} Inspection`}
            description={`Registered Customer: ${selCustomer}`}
            maxWidth="lg"
          >
            <div className="space-y-6 text-left">
              {/* Visual Lifecycle Progress */}
              <div>
                <p className="text-xs font-semibold uppercase tracking-wider text-muted-foreground mb-1">
                  Fulfillment Progress
                </p>
                <div className="p-4 bg-surface rounded-xl border border-border">
                  <OrderProgress status={selStatus} />
                </div>
              </div>

              {/* Items Breakdown */}
              <div>
                <p className="text-xs font-semibold uppercase tracking-wider text-muted-foreground mb-2">
                  Order Items
                </p>
                <div className="border border-border rounded-xl divide-y divide-border overflow-hidden">
                  {selItems.map((item: any, idx: number) => {
                    const itemId = item.ID || item.id || '';
                    const itemName = item.Name || item.name || '';
                    const itemQty = item.Quantity || item.quantity || 1;
                    const meta = getProductMeta(itemId, itemName);
                    return (
                      <div key={idx} className="p-3 flex items-center justify-between text-xs">
                        <div className="flex items-center gap-3">
                          <img
                            src={meta.image}
                            alt={meta.name}
                            className="w-10 h-10 rounded object-cover border border-border"
                          />
                          <div>
                            <p className="font-bold text-foreground">{meta.name}</p>
                            <p className="text-muted-foreground font-mono">
                              ID: {itemId} · Qty: {itemQty}
                            </p>
                          </div>
                        </div>
                        <span className="font-mono font-bold text-foreground">
                          {formatCurrency(meta.price * itemQty)}
                        </span>
                      </div>
                    );
                  })}
                </div>
              </div>

              {/* Order Fulfillment Details */}
              <div className="p-3.5 bg-surface rounded-xl border border-border space-y-1.5 text-xs text-muted-foreground">
                <div className="flex justify-between">
                  <span>Fulfillment Type:</span>
                  <span className="font-medium text-foreground">Kitchen Order</span>
                </div>
                <div className="flex justify-between">
                  <span>Payment Session:</span>
                  <span className="font-medium text-foreground">{selPaymentLink ? 'Generated' : 'Pending Generation'}</span>
                </div>
                <div className="flex justify-between">
                  <span>Fulfillment Status:</span>
                  <span className="font-medium text-emerald-600 capitalize">{selStatus}</span>
                </div>
              </div>

              <div className="flex justify-between items-center pt-2">
                {selStatus !== 'ready' && selStatus !== 'cancelled' && selStatus !== 'refunded' ? (
                  <Button
                    onClick={() => {
                      if (window.confirm(`Initiate Saga cancellation for Order #${selId}? This will automatically revert stock and refund customer payment.`)) {
                        adminCancelMutation.mutate(selId, {
                          onSuccess: () => setSelectedOrder(null),
                        });
                      }
                    }}
                    variant="danger"
                    size="sm"
                    disabled={adminCancelMutation.isPending}
                    leftIcon={<XCircle className="w-3.5 h-3.5" />}
                  >
                    {adminCancelMutation.isPending ? 'Cancelling...' : 'Cancel & Compensate'}
                  </Button>
                ) : <div />}
                <Button onClick={() => setSelectedOrder(null)} variant="primary" size="sm">
                  Close
                </Button>
              </div>
            </div>
          </Dialog>
        );
      })()}
    </div>
  );
}
