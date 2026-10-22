import React, { useState, useEffect } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import { useOrderDetail, useVerifyPayment, useCancelOrder } from '@/features/orders/useOrders';
import { useAuthStore } from '@/features/auth/useAuthStore';
import { getProductMeta } from '@/config/site';
import { formatCurrency } from '@/lib/utils';
import { Button } from '@/components/ui/Button';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/Card';
import { Badge } from '@/components/ui/Badge';
import { OrderProgress } from '@/components/ui/OrderProgress';
import { Skeleton } from '@/components/ui/Skeleton';
import { EmptyState } from '@/components/ui/EmptyState';
import { toast } from 'sonner';
import {
  ArrowLeft,
  CreditCard,
  ExternalLink,
  RefreshCw,
  Clock,
  Sparkles,
  ChefHat,
  CheckCircle,
  CheckCircle2,
  XCircle,
} from 'lucide-react';

export function OrderDetailPage() {
  const { orderId } = useParams<{ orderId: string }>();
  const navigate = useNavigate();
  const { customerID } = useAuthStore();
  const { data: order, isLoading, isError, refetch, isFetching } = useOrderDetail(orderId);
  const verifyMutation = useVerifyPayment();
  const cancelMutation = useCancelOrder();
  const [isVerifying, setIsVerifying] = useState(false);

  const status = (order?.Status || 'pending').toLowerCase();
  const isPaid = status === 'paid';
  const isReady = status === 'ready';
  const isCancelled = status === 'cancelled' || status === 'refunded';
  const paymentLink = order?.PaymentLink;
  const isMockPayment =
    paymentLink &&
    (paymentLink.includes('success.html') || paymentLink.startsWith('/'));

  const handleCancelOrder = async () => {
    if (!order?.ID) return;
    if (!window.confirm('Cancel this order? This will execute the cancellation Saga to refund payment and restore inventory.')) {
      return;
    }
    try {
      await cancelMutation.mutateAsync(order.ID);
      refetch();
    } catch (err) {
      // toast shown in mutation
    }
  };


  const handleVerifyPayment = async () => {
    if (!order?.ID) return;
    setIsVerifying(true);
    try {
      const res = await verifyMutation.mutateAsync(order.ID);
      if (res.verified) {
        toast.success('Payment verified! Your order is now being prepared by the kitchen.');
        refetch();
      } else {
        toast.info(res.message || 'Payment has not been completed on Razorpay yet.');
      }
    } catch (err: any) {
      toast.error(err.message || 'Failed to verify payment status.');
    } finally {
      setIsVerifying(false);
    }
  };

  // Automatically check payment when user returns to this browser tab (Hook called unconditionally)
  useEffect(() => {
    const onFocus = () => {
      if (order?.ID && !isPaid && !isReady) {
        verifyMutation.mutateAsync(order.ID)
          .then((res) => {
            if (res.verified) {
              toast.success('Payment confirmed by Razorpay!');
              refetch();
            }
          })
          .catch(() => {});
      }
    };
    window.addEventListener('focus', onFocus);
    return () => window.removeEventListener('focus', onFocus);
  }, [order?.ID, isPaid, isReady]);

  if (isLoading) {
    return (
      <div className="max-w-4xl mx-auto px-4 py-12 space-y-6">
        <Skeleton className="h-6 w-32" />
        <Skeleton className="h-10 w-64" />
        <Skeleton className="h-48 w-full rounded-xl" />
        <Skeleton className="h-64 w-full rounded-xl" />
      </div>
    );
  }

  if (isError || !order) {
    return (
      <div className="max-w-2xl mx-auto px-4 py-16">
        <EmptyState
          icon={<Clock className="w-6 h-6" />}
          title="Order Not Found"
          description="Could not locate this order under your active customer session."
          actionLabel="View All Orders"
          onAction={() => navigate('/orders')}
        />
      </div>
    );
  }

  const items = order.Items || [];
  const orderTotal = items.reduce((acc, item) => {
    const meta = getProductMeta(item.ID);
    return acc + meta.price * (item.Quantity || 1);
  }, 0);

  return (
    <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-8 text-left">
      {/* Back Link */}
      <div className="flex items-center justify-between">
        <Link
          to="/orders"
          className="inline-flex items-center gap-1.5 text-xs font-semibold text-muted-foreground hover:text-foreground transition-colors"
        >
          <ArrowLeft className="w-4 h-4" />
          <span>Back to All Orders</span>
        </Link>
        <Button
          onClick={() => refetch()}
          disabled={isFetching}
          variant="ghost"
          size="sm"
          leftIcon={<RefreshCw className={`w-3.5 h-3.5 ${isFetching ? 'animate-spin' : ''}`} />}
        >
          {isFetching ? 'Updating...' : 'Live Refresh'}
        </Button>
      </div>

      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-4 border-b border-border">
        <div>
          <span className="text-xs font-mono uppercase tracking-wider text-muted-foreground">
            Customer Order
          </span>
          <h1 className="text-3xl font-bold font-editorial text-foreground">
            Order #{order.ID}
          </h1>
          <p className="text-xs text-muted-foreground font-mono mt-1">
            Registered Identity: <span className="text-foreground">{order.customerID || customerID}</span>
          </p>
        </div>
        <div className="flex items-center gap-3">
          {!isReady && !isCancelled && (
            <Button
              onClick={handleCancelOrder}
              disabled={cancelMutation.isPending}
              variant="outline"
              size="sm"
              className="text-rose-500 hover:text-rose-600 hover:bg-rose-500/10 border-rose-500/20"
              leftIcon={<XCircle className="w-3.5 h-3.5" />}
            >
              {cancelMutation.isPending ? 'Cancelling...' : 'Cancel Order'}
            </Button>
          )}
          <Badge
            variant={isCancelled ? 'danger' : isReady ? 'ready' : isPaid ? 'paid' : 'pending'}
            size="md"
            className="self-start sm:self-auto text-sm px-3.5 py-1"
          >
            Status: {order.Status || 'Pending'}
          </Badge>
        </div>
      </div>

      {/* Visual Order Lifecycle */}
      <Card className="p-6 border border-border shadow-subtle">
        <OrderProgress status={order.Status} />
      </Card>

      {/* Payment Action Bar (If Pending) */}
      {!isPaid && !isReady && (
        <Card className="p-6 border border-border shadow-elevated bg-surface/50 space-y-4">
          <div className="flex items-start justify-between gap-4">
            <div>
              <h3 className="font-bold text-base text-foreground">Awaiting Payment Confirmation</h3>
              <p className="text-xs text-muted-foreground mt-0.5">
                Your order has been received. Please complete payment to send your order to the kitchen for preparation.
              </p>
            </div>
            <Clock className="w-5 h-5 text-amber-500 animate-pulse shrink-0" />
          </div>

          {isMockPayment && (
            <div className="p-3.5 rounded-lg bg-amber-500/10 border border-amber-500/20 text-amber-800 dark:text-amber-300 text-xs flex items-center gap-2">
              <Sparkles className="w-4 h-4 shrink-0" />
              <span>Test Payment Mode Active: Instant payment simulation available.</span>
            </div>
          )}

          {paymentLink ? (
            <div className="flex flex-col sm:flex-row items-stretch sm:items-center gap-3 pt-1">
              <a
                href={paymentLink}
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex items-center justify-center font-medium rounded-lg text-sm px-6 py-3 gap-2 bg-emerald-600 hover:bg-emerald-700 text-white shadow-sm transition-all"
              >
                <CreditCard className="w-4 h-4" />
                <span>Complete Payment via Gateway</span>
                <ExternalLink className="w-4 h-4" />
              </a>

              <Button
                onClick={handleVerifyPayment}
                disabled={isVerifying}
                variant="outline"
                size="md"
                className="border-emerald-600/30 text-emerald-700 dark:text-emerald-400 hover:bg-emerald-500/10"
                leftIcon={<RefreshCw className={`w-4 h-4 ${isVerifying ? 'animate-spin' : ''}`} />}
              >
                {isVerifying ? 'Verifying payment...' : 'Verify Payment Status'}
              </Button>
            </div>
          ) : (
            <div className="text-xs text-muted-foreground italic flex items-center gap-2">
              <Clock className="w-3.5 h-3.5 animate-spin" />
              <span>Generating secure payment link...</span>
            </div>
          )}
        </Card>
      )}

      {/* Preparation Status (If Paid) */}
      {isPaid && !isReady && (
        <div className="p-5 rounded-xl bg-emerald-500/10 border border-emerald-500/20 text-emerald-800 dark:text-emerald-300 flex items-center gap-3">
          <ChefHat className="w-6 h-6 shrink-0 text-emerald-600 animate-bounce" />
          <div>
            <p className="font-bold text-sm">Payment Verified — Kitchen Preparing Order</p>
            <p className="text-xs mt-0.5">
              Payment confirmed! Our chefs have received your ticket and are preparing your meal.
            </p>
          </div>
        </div>
      )}

      {/* Ready Status */}
      {isReady && (
        <div className="p-5 rounded-xl bg-blue-500/10 border border-blue-500/20 text-blue-800 dark:text-blue-300 flex items-center gap-3">
          <CheckCircle className="w-6 h-6 shrink-0 text-blue-600" />
          <div>
            <p className="font-bold text-sm">Order Prepared & Ready</p>
            <p className="text-xs mt-0.5">
              Your meal is freshly prepared and ready for pickup or delivery!
            </p>
          </div>
        </div>
      )}

      {/* Items Breakdown Table */}
      <Card className="border border-border overflow-hidden shadow-subtle">
        <CardHeader className="bg-surface/50 border-b border-border/60 p-4 sm:p-6">
          <CardTitle className="text-base font-bold">Order Items ({items.length})</CardTitle>
        </CardHeader>
        <div className="divide-y divide-border">
          {items.map((item, idx) => {
            const meta = getProductMeta(item.ID, item.Name);
            const lineTotal = meta.price * (item.Quantity || 1);

            return (
              <div key={idx} className="p-4 sm:p-5 flex items-center justify-between gap-4">
                <div className="flex items-center gap-4">
                  <img
                    src={meta.image}
                    alt={meta.name}
                    className="w-14 h-14 rounded-lg object-cover border border-border bg-muted"
                  />
                  <div>
                    <p className="font-bold text-sm text-foreground">{meta.name}</p>
                    <p className="text-xs text-muted-foreground font-mono">
                      Quantity: {item.Quantity} × {formatCurrency(meta.price)}
                    </p>
                  </div>
                </div>
                <div className="text-right font-mono font-bold text-sm text-foreground">
                  {formatCurrency(lineTotal)}
                </div>
              </div>
            );
          })}
        </div>
        <div className="p-5 bg-surface/30 border-t border-border flex justify-between items-baseline text-sm">
          <span className="font-bold text-foreground">Estimated Order Total</span>
          <span className="font-mono font-extrabold text-xl text-primary">
            {formatCurrency(orderTotal)}
          </span>
        </div>
      </Card>
    </div>
  );
}
