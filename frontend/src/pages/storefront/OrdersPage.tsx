import React from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useCustomerOrders } from '@/features/orders/useOrders';
import { useAuthStore } from '@/features/auth/useAuthStore';
import { getProductMeta } from '@/config/site';
import { formatCurrency } from '@/lib/utils';
import { Button } from '@/components/ui/Button';
import { Card } from '@/components/ui/Card';
import { Badge } from '@/components/ui/Badge';
import { Skeleton } from '@/components/ui/Skeleton';
import { EmptyState } from '@/components/ui/EmptyState';
import { Package, ArrowRight, Clock, RefreshCw } from 'lucide-react';
import { Order } from '@/types';

export function OrdersPage() {
  const navigate = useNavigate();
  const { data: orders, isLoading, isError, refetch } = useCustomerOrders();
  const { isAuthenticated, customerID, isAdmin } = useAuthStore();

  if (isAdmin) {
    return (
      <div className="max-w-2xl mx-auto px-4 py-16">
        <EmptyState
          icon={<Package className="w-6 h-6" />}
          title="Customer Orders Portal"
          description="You are currently signed in with administrator privileges. All platform customer orders are managed in the Order Operations console."
          actionLabel="View Admin Orders"
          onAction={() => navigate('/admin/orders')}
        />
      </div>
    );
  }

  if (!isAuthenticated) {
    return (
      <div className="max-w-2xl mx-auto px-4 py-16">
        <EmptyState
          icon={<Package className="w-6 h-6" />}
          title="Sign in to View Order History"
          description="Your orders are securely associated with your authenticated identity."
          actionLabel="Sign In"
          onAction={() => navigate('/login')}
        />
      </div>
    );
  }

  return (
    <div className="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-8 text-left">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between pb-4 border-b border-border gap-4">
        <div>
          <h1 className="text-3xl font-bold font-editorial tracking-tight text-foreground">
            Order History
          </h1>
          <p className="text-sm text-muted-foreground">
            Showing all orders for <span className="font-mono text-foreground font-semibold">{customerID}</span>
          </p>
        </div>
        <Button
          onClick={() => refetch()}
          variant="ghost"
          size="sm"
          className="self-start sm:self-auto"
          leftIcon={<RefreshCw className="w-3.5 h-3.5" />}
        >
          Refresh
        </Button>
      </div>

      {isLoading ? (
        <div className="space-y-4">
          {Array.from({ length: 3 }).map((_, i) => (
            <Card key={i} className="p-6 border border-border space-y-3">
              <div className="flex justify-between">
                <Skeleton className="h-5 w-32" />
                <Skeleton className="h-5 w-20" />
              </div>
              <Skeleton className="h-4 w-48" />
              <Skeleton className="h-10 w-full" />
            </Card>
          ))}
        </div>
      ) : isError ? (
        <EmptyState
          icon={<Clock className="w-6 h-6" />}
          title="Could Not Load Orders"
          description="Unable to load orders at this moment. Please try again in a few moments."
          actionLabel="Try Again"
          onAction={() => refetch()}
        />
      ) : !orders || orders.length === 0 ? (
        <EmptyState
          icon={<Package className="w-6 h-6" />}
          title="No Orders Yet"
          description="You haven't placed any orders with STRATUM yet. Explore our handcrafted provisions menu to place your first order."
          actionLabel="Browse Catalog"
          onAction={() => navigate('/products')}
        />
      ) : (
        <div className="space-y-4">
          {orders.map((order: Order) => {
            const status = (order.Status || 'pending').toLowerCase();
            const badgeVariant =
              status === 'ready'
                ? 'ready'
                : status === 'paid'
                ? 'paid'
                : 'pending';

            const items = order.Items || [];
            const totalItemsCount = items.reduce(
              (acc, item) => acc + (item.Quantity || 0),
              0
            );

            // Compute approximate total based on catalog pricing
            const orderTotal = items.reduce((acc, item) => {
              const meta = getProductMeta(item.ID);
              return acc + meta.price * (item.Quantity || 1);
            }, 0);

            return (
              <Card
                key={order.ID}
                className="p-6 border border-border shadow-subtle hover:border-border/80 transition-all flex flex-col sm:flex-row sm:items-center justify-between gap-6"
              >
                <div className="space-y-3">
                  <div className="flex items-center gap-3">
                    <span className="font-mono text-base font-bold text-foreground">
                      Order #{order.ID}
                    </span>
                    <Badge variant={badgeVariant} size="sm">
                      {order.Status || 'Pending'}
                    </Badge>
                  </div>

                  {/* Items Summary preview */}
                  <div className="flex flex-wrap gap-2 text-xs text-muted-foreground">
                    {items.map((item, idx) => {
                      const meta = getProductMeta(item.ID, item.Name);
                      return (
                        <span
                          key={idx}
                          className="bg-surface px-2.5 py-1 rounded-md border border-border/60"
                        >
                          {item.Quantity}× {meta.name}
                        </span>
                      );
                    })}
                  </div>

                  <div className="text-xs text-muted-foreground font-mono">
                    Total: <span className="font-bold text-foreground">{formatCurrency(orderTotal)}</span> · {totalItemsCount} {totalItemsCount === 1 ? 'item' : 'items'}
                  </div>
                </div>

                <div className="flex items-center gap-3 shrink-0">
                  <Button
                    onClick={() => navigate(`/orders/${order.ID}`)}
                    variant="outline"
                    size="sm"
                    rightIcon={<ArrowRight className="w-3.5 h-3.5" />}
                  >
                    Track Status
                  </Button>
                </div>
              </Card>
            );
          })}
        </div>
      )}
    </div>
  );
}
