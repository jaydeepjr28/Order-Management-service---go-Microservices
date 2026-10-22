import React, { useState, useEffect } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useCart } from '@/features/cart/useCart';
import { useAuthStore } from '@/features/auth/useAuthStore';
import { useCreateOrder, useOrderDetail } from '@/features/orders/useOrders';
import { formatCurrency } from '@/lib/utils';
import { Button } from '@/components/ui/Button';
import { Card, CardHeader, CardTitle, CardContent, CardFooter } from '@/components/ui/Card';
import { Badge } from '@/components/ui/Badge';
import { OrderProgress } from '@/components/ui/OrderProgress';
import {
  ShieldCheck,
  CreditCard,
  ExternalLink,
  Clock,
  ArrowRight,
  AlertCircle,
  CheckCircle,
  Sparkles,
} from 'lucide-react';

export function CheckoutPage() {
  const navigate = useNavigate();
  const { items, totalQuantity, totalAmount, clearCart } = useCart();
  const { customerID, isAuthenticated, isAdmin } = useAuthStore();
  const createOrderMutation = useCreateOrder();

  const [activeOrderId, setActiveOrderId] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const isSubmittingRef = React.useRef(false);

  // Poll for the created order to receive the asynchronous PaymentLink
  const orderDetailQuery = useOrderDetail(activeOrderId || undefined);
  const currentOrder = orderDetailQuery.data;

  // If user is unauthenticated or admin
  useEffect(() => {
    if (!isAuthenticated) {
      navigate('/login?redirect=/checkout');
    } else if (isAdmin) {
      navigate('/admin');
    }
  }, [isAuthenticated, isAdmin, navigate]);

  const handlePlaceOrder = async () => {
    if (
      isAdmin ||
      isSubmittingRef.current ||
      isSubmitting ||
      createOrderMutation.isPending ||
      items.length === 0 ||
      !customerID
    ) {
      return;
    }
    isSubmittingRef.current = true;
    setIsSubmitting(true);

    try {
      const payload = items.map((item) => ({
        ID: item.productId,
        Quantity: item.quantity,
      }));

      const res = await createOrderMutation.mutateAsync(payload);
      const newOrderId = res?.order?.ID || (res?.order as any)?.id || (res as any)?.ID || (res as any)?.id;

      if (newOrderId) {
        setActiveOrderId(newOrderId);
        // Clear cart now that order has been accepted
        clearCart();
      }
    } catch (err) {
      console.error('Order creation failed:', err);
    } finally {
      setIsSubmitting(false);
      isSubmittingRef.current = false;
    }
  };

  // State 2: Order has been placed; now waiting for / handling payment link
  if (activeOrderId) {
    const paymentLink = currentOrder?.PaymentLink;
    const isMockPayment =
      paymentLink &&
      (paymentLink.includes('success.html') || paymentLink.startsWith('/'));
    const isPaid = currentOrder?.Status?.toLowerCase() === 'paid';
    const isReady = currentOrder?.Status?.toLowerCase() === 'ready';

    return (
      <div className="max-w-3xl mx-auto px-4 sm:px-6 py-12 space-y-8 text-left">
        {/* Progress Header */}
        <div className="space-y-2">
          <Badge variant="primary" size="sm">
            Order #{activeOrderId}
          </Badge>
          <h1 className="text-3xl font-bold font-editorial text-foreground">
            {isReady
              ? 'Your Order is Ready!'
              : isPaid
              ? 'Payment Confirmed — In Preparation'
              : 'Order Created — Awaiting Payment'}
          </h1>
          <p className="text-sm text-muted-foreground">
            Your order has been initiated. Complete payment below to confirm.
          </p>
        </div>

        {/* Visual Lifecycle Progress */}
        <Card className="p-6 border border-border shadow-subtle">
          <OrderProgress status={currentOrder?.Status || 'pending'} />
        </Card>

        {/* Payment Action Box */}
        <Card className="border border-border shadow-elevated p-6 space-y-6">
          <div className="flex items-start justify-between gap-4">
            <div>
              <h3 className="text-base font-bold text-foreground">Payment Status</h3>
              <p className="text-xs text-muted-foreground mt-0.5">
                {isPaid
                  ? 'Payment has been successfully verified.'
                  : paymentLink
                  ? 'Payment gateway link generated. Complete checkout to confirm your order.'
                  : 'Generating secure checkout link...'}
              </p>
            </div>
            <Badge
              variant={isReady ? 'ready' : isPaid ? 'paid' : 'pending'}
              size="md"
            >
              {currentOrder?.Status || 'Pending'}
            </Badge>
          </div>

          {/* Test Mode Notice */}
          {isMockPayment && !isPaid && (
            <div className="p-4 rounded-xl bg-amber-500/10 border border-amber-500/20 text-amber-800 dark:text-amber-300 text-xs space-y-1">
              <div className="flex items-center gap-1.5 font-bold">
                <Sparkles className="w-4 h-4" />
                <span>Test Mode Active</span>
              </div>
              <p className="leading-relaxed">
                Click below to simulate instant payment verification.
              </p>
            </div>
          )}

          {/* Payment Link Button */}
          {!isPaid && !isReady ? (
            paymentLink ? (
              <div className="space-y-3">
                <a
                  href={paymentLink}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="w-full inline-flex items-center justify-center font-medium rounded-lg text-sm px-6 py-3.5 gap-2 bg-emerald-600 hover:bg-emerald-700 text-white shadow-sm transition-all"
                >
                  <CreditCard className="w-4 h-4" />
                  <span>Continue to Payment Gateway</span>
                  <ExternalLink className="w-4 h-4" />
                </a>
                <p className="text-[11px] text-center text-muted-foreground">
                  The order status will automatically update once payment is complete.
                </p>
              </div>
            ) : (
              <div className="flex items-center justify-center p-6 bg-surface rounded-xl border border-dashed border-border gap-3 text-xs text-muted-foreground">
                <Clock className="w-4 h-4 animate-spin text-primary" />
                <span>Preparing payment details...</span>
              </div>
            )
          ) : (
            <div className="p-4 rounded-xl bg-emerald-500/10 border border-emerald-500/20 text-emerald-800 dark:text-emerald-300 text-xs flex items-center gap-2">
              <CheckCircle className="w-4 h-4 shrink-0" />
              <span>Payment confirmed! Our chefs are preparing your order.</span>
            </div>
          )}

          <div className="pt-4 border-t border-border flex justify-between items-center text-xs">
            <Link
              to={`/orders/${activeOrderId}`}
              className="text-primary hover:underline font-semibold flex items-center gap-1"
            >
              <span>View full order details & live tracking</span>
              <ArrowRight className="w-3.5 h-3.5" />
            </Link>
            <Button onClick={() => navigate('/products')} variant="ghost" size="sm">
              Back to Catalog
            </Button>
          </div>
        </Card>
      </div>
    );
  }

  // State 1: Review cart and place order
  if (items.length === 0) {
    return (
      <div className="max-w-2xl mx-auto px-4 py-16 text-center space-y-4">
        <h2 className="text-xl font-bold text-foreground">No Items to Checkout</h2>
        <p className="text-sm text-muted-foreground">
          Your shopping cart is currently empty. Add products to proceed with checkout.
        </p>
        <Button onClick={() => navigate('/products')} variant="primary" size="md">
          Return to Menu
        </Button>
      </div>
    );
  }

  return (
    <div className="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-8 text-left">
      <div>
        <h1 className="text-3xl font-bold font-editorial tracking-tight text-foreground">
          Checkout Confirmation
        </h1>
        <p className="text-sm text-muted-foreground mt-1">
          Review your order details and items before placing your order.
        </p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 items-start">
        {/* Order Items Table */}
        <div className="lg:col-span-7 space-y-4">
          <Card className="p-6 border border-border shadow-subtle space-y-4">
            <h3 className="text-base font-bold text-foreground">Items in this Order</h3>
            <div className="divide-y divide-border">
              {items.map((item) => (
                <div key={item.productId} className="py-3 flex items-center justify-between gap-4">
                  <div className="flex items-center gap-3">
                    <img
                      src={item.image}
                      alt={item.name}
                      className="w-12 h-12 rounded-lg object-cover border border-border bg-muted"
                    />
                    <div>
                      <p className="font-semibold text-sm text-foreground">{item.name}</p>
                      <p className="text-xs text-muted-foreground font-mono">
                        {item.quantity} × {formatCurrency(item.price)}
                      </p>
                    </div>
                  </div>
                  <span className="font-mono font-bold text-sm text-foreground">
                    {formatCurrency(item.totalPrice)}
                  </span>
                </div>
              ))}
            </div>
          </Card>

          {/* Customer Account Info */}
          <Card className="p-6 border border-border shadow-subtle space-y-2">
            <h3 className="text-base font-bold text-foreground">Customer Profile</h3>
            <p className="text-xs text-muted-foreground">
              Order will be assigned to authenticated identity:
            </p>
            <p className="font-mono text-sm font-semibold text-primary">{customerID}</p>
          </Card>
        </div>

        {/* Order Payment Summary */}
        <div className="lg:col-span-5">
          <Card className="p-6 border border-border shadow-elevated space-y-5">
            <h3 className="text-base font-bold text-foreground">Payment Summary</h3>

            <div className="space-y-3 text-sm">
              <div className="flex justify-between text-muted-foreground">
                <span>Subtotal ({totalQuantity} items)</span>
                <span className="font-mono text-foreground">{formatCurrency(totalAmount)}</span>
              </div>
              <div className="flex justify-between text-muted-foreground">
                <span>Estimated Packaging</span>
                <span className="font-mono text-emerald-600 font-medium">Free</span>
              </div>
              <div className="pt-3 border-t border-border flex justify-between items-baseline">
                <span className="font-bold text-base text-foreground">Total to Pay</span>
                <span className="font-extrabold text-2xl font-mono text-primary">
                  {formatCurrency(totalAmount)}
                </span>
              </div>
            </div>

            <Button
              onClick={handlePlaceOrder}
              isLoading={isSubmitting || createOrderMutation.isPending}
              variant="primary"
              size="lg"
              className="w-full"
            >
              Confirm & Submit Order
            </Button>

            <div className="space-y-2 pt-2 border-t border-border text-[11px] text-muted-foreground">
              <div className="flex items-center gap-1.5">
                <ShieldCheck className="w-3.5 h-3.5 text-emerald-600" />
                <span>Real-time inventory reservation for your items</span>
              </div>
              <div className="flex items-center gap-1.5">
                <CreditCard className="w-3.5 h-3.5 text-emerald-600" />
                <span>Secure 256-bit encrypted payment processing</span>
              </div>
            </div>
          </Card>
        </div>
      </div>
    </div>
  );
}
