import React, { useEffect } from 'react';
import { useSearchParams, useNavigate, Link } from 'react-router-dom';
import { useOrderDetail, useVerifyPayment } from '@/features/orders/useOrders';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { OrderProgress } from '@/components/ui/OrderProgress';
import { CheckCircle2, ArrowRight, ShoppingBag, Clock } from 'lucide-react';
import { toast } from 'sonner';

export function PaymentSuccessPage() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const orderId = searchParams.get('orderID') || searchParams.get('orderId') || '';
  const customerId = searchParams.get('customerID') || searchParams.get('customerId') || '';

  const { data: order, isLoading, refetch } = useOrderDetail(orderId || undefined);
  const verifyMutation = useVerifyPayment();

  useEffect(() => {
    if (orderId) {
      verifyMutation.mutateAsync(orderId)
        .then((res) => {
          if (res.verified) {
            toast.success('Payment verified! Your order is now being prepared by the kitchen.');
            refetch();
          }
        })
        .catch(() => {});
    }
  }, [orderId]);

  return (
    <div className="max-w-2xl mx-auto px-4 py-16 text-left space-y-6">
      <Card className="p-8 border border-border shadow-elevated text-center space-y-6">
        <div className="w-16 h-16 rounded-full bg-emerald-500/10 text-emerald-600 flex items-center justify-center mx-auto">
          <CheckCircle2 className="w-8 h-8" />
        </div>

        <div className="space-y-2">
          <Badge variant="paid" size="sm">
            Payment Completed
          </Badge>
          <h1 className="text-3xl font-bold font-editorial text-foreground">
            Thank You for Your Order!
          </h1>
          <p className="text-sm text-muted-foreground max-w-md mx-auto">
            Your transaction has been processed and your order has been received by our kitchen.
          </p>
        </div>

        {orderId && (
          <div className="p-4 bg-surface rounded-xl border border-border text-left space-y-3">
            <div className="flex items-center justify-between text-xs font-mono">
              <span className="text-muted-foreground">Order Reference:</span>
              <span className="font-bold text-foreground">#{orderId}</span>
            </div>
            {customerId && (
              <div className="flex items-center justify-between text-xs font-mono">
                <span className="text-muted-foreground">Customer:</span>
                <span className="font-medium text-foreground">{customerId}</span>
              </div>
            )}
            <div className="flex items-center justify-between text-xs font-mono">
              <span className="text-muted-foreground">Current Status:</span>
              <Badge variant={order?.Status === 'ready' ? 'ready' : 'paid'} size="sm">
                {order?.Status || 'Paid / In Preparation'}
              </Badge>
            </div>
          </div>
        )}

        {/* Visual Lifecycle Progress */}
        <div className="p-4 bg-surface/50 rounded-xl border border-border text-left">
          <p className="text-xs font-semibold uppercase tracking-wider text-muted-foreground mb-3">
            Fulfillment Lifecycle
          </p>
          <OrderProgress status={order?.Status || 'paid'} />
        </div>

        <div className="flex flex-col sm:flex-row items-center justify-center gap-3 pt-4 border-t border-border">
          {orderId && (
            <Button
              onClick={() => navigate(`/orders/${orderId}`)}
              variant="primary"
              size="md"
              rightIcon={<ArrowRight className="w-4 h-4" />}
            >
              Track Order Live
            </Button>
          )}
          <Button
            onClick={() => navigate('/products')}
            variant="outline"
            size="md"
            leftIcon={<ShoppingBag className="w-4 h-4" />}
          >
            Return to Menu
          </Button>
        </div>
      </Card>
    </div>
  );
}
