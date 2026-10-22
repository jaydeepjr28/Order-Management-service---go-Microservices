import React from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useCart } from '@/features/cart/useCart';
import { useAuthStore } from '@/features/auth/useAuthStore';
import { formatCurrency } from '@/lib/utils';
import { Button } from '@/components/ui/Button';
import { Card, CardHeader, CardTitle, CardContent, CardFooter } from '@/components/ui/Card';
import { EmptyState } from '@/components/ui/EmptyState';
import {
  ShoppingBag,
  Trash2,
  ArrowRight,
  ShieldCheck,
  RotateCcw,
} from 'lucide-react';

export function CartPage() {
  const navigate = useNavigate();
  const { items, totalQuantity, totalAmount, removeItem, isRemoving, clearCart, isClearing } =
    useCart();
  const { isAuthenticated, isAdmin } = useAuthStore();

  if (isAdmin) {
    return (
      <div className="max-w-2xl mx-auto px-4 py-16">
        <EmptyState
          icon={<ShieldCheck className="w-6 h-6 text-primary" />}
          title="Shopping Restricted for Administrators"
          description="Administrator accounts are strictly dedicated to catalog, fulfillment, and operations management. Please sign in with a customer account to purchase provisions."
          actionLabel="Go to Admin Console"
          onAction={() => navigate('/admin')}
        />
      </div>
    );
  }

  if (!isAuthenticated) {
    return (
      <div className="max-w-2xl mx-auto px-4 py-16">
        <EmptyState
          icon={<ShoppingBag className="w-6 h-6" />}
          title="Sign in to View Your Cart"
          description="Sign in to view your saved cart items and proceed to checkout."
          actionLabel="Sign In Now"
          onAction={() => navigate('/login')}
        />
      </div>
    );
  }

  if (items.length === 0) {
    return (
      <div className="max-w-2xl mx-auto px-4 py-16">
        <EmptyState
          icon={<ShoppingBag className="w-6 h-6" />}
          title="Your Cart is Empty"
          description="You haven't added any provisions to your cart yet. Explore our handcrafted menu to get started."
          actionLabel="Explore Menu"
          onAction={() => navigate('/products')}
        />
      </div>
    );
  }

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-8">
      {/* Title */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between pb-4 border-b border-border gap-4">
        <div>
          <h1 className="text-3xl font-bold font-editorial tracking-tight text-foreground">
            Shopping Cart
          </h1>
          <p className="text-sm text-muted-foreground">
            {totalQuantity} {totalQuantity === 1 ? 'item' : 'items'} in your active cart session.
          </p>
        </div>
        <Button
          onClick={() => clearCart()}
          disabled={isClearing}
          variant="ghost"
          size="sm"
          className="text-muted-foreground hover:text-red-500 self-start sm:self-auto"
          leftIcon={<RotateCcw className="w-3.5 h-3.5" />}
        >
          Clear Cart
        </Button>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 items-start">
        {/* Items List */}
        <div className="lg:col-span-8 space-y-4">
          {items.map((item) => (
            <div
              key={item.productId}
              className="p-4 sm:p-5 rounded-xl border border-border bg-card flex flex-col sm:flex-row items-center gap-4 sm:gap-6 shadow-subtle hover:border-border/80 transition-all"
            >
              <div className="w-20 h-20 rounded-lg overflow-hidden border border-border bg-muted shrink-0">
                <img
                  src={item.image}
                  alt={item.name}
                  className="w-full h-full object-cover"
                />
              </div>

              <div className="flex-1 text-center sm:text-left space-y-1 min-w-0">
                <Link
                  to={`/products/${item.productId}`}
                  className="font-bold text-base text-foreground hover:text-primary transition-colors block truncate"
                >
                  {item.name}
                </Link>
                <p className="text-xs text-muted-foreground font-mono">
                  Unit: {formatCurrency(item.price)}
                </p>
              </div>

              <div className="flex items-center gap-6 sm:gap-8">
                <div className="text-center">
                  <span className="text-[10px] uppercase font-semibold text-muted-foreground block">
                    Qty
                  </span>
                  <span className="font-bold text-sm font-mono text-foreground">
                    {item.quantity}
                  </span>
                </div>

                <div className="text-right min-w-[80px]">
                  <span className="text-[10px] uppercase font-semibold text-muted-foreground block">
                    Total
                  </span>
                  <span className="font-extrabold text-base font-mono text-foreground">
                    {formatCurrency(item.totalPrice)}
                  </span>
                </div>

                <button
                  onClick={() => removeItem(item.productId)}
                  disabled={isRemoving}
                  className="p-2 text-muted-foreground hover:text-red-500 rounded-lg hover:bg-surface transition-colors"
                  aria-label={`Remove ${item.name}`}
                >
                  <Trash2 className="w-4 h-4" />
                </button>
              </div>
            </div>
          ))}
        </div>

        {/* Order Summary Column */}
        <div className="lg:col-span-4">
          <Card className="sticky top-24 border border-border shadow-elevated">
            <CardHeader className="border-b border-border/60 pb-4">
              <CardTitle className="text-base font-bold">Order Summary</CardTitle>
            </CardHeader>
            <CardContent className="pt-5 space-y-4 text-sm">
              <div className="flex justify-between text-muted-foreground">
                <span>Subtotal ({totalQuantity} items)</span>
                <span className="font-mono text-foreground">{formatCurrency(totalAmount)}</span>
              </div>
              <div className="flex justify-between text-muted-foreground">
                <span>Estimated Packaging</span>
                <span className="font-mono text-emerald-600 font-medium">Free</span>
              </div>
              <div className="flex justify-between text-muted-foreground">
                <span>Taxes</span>
                <span className="text-xs text-muted-foreground">Included</span>
              </div>

              <div className="pt-4 border-t border-border flex justify-between items-baseline">
                <span className="font-bold text-base text-foreground">Total</span>
                <span className="font-extrabold text-2xl font-mono text-primary">
                  {formatCurrency(totalAmount)}
                </span>
              </div>
            </CardContent>
            <CardFooter className="pt-2 flex flex-col gap-3">
              <Button
                onClick={() => navigate('/checkout')}
                variant="primary"
                size="lg"
                className="w-full"
                rightIcon={<ArrowRight className="w-4 h-4" />}
              >
                Proceed to Checkout
              </Button>
              <div className="flex items-center justify-center gap-1.5 text-[11px] text-muted-foreground">
                <ShieldCheck className="w-3.5 h-3.5 text-emerald-600" />
                <span>Verified Stock Check Prior to Confirmation</span>
              </div>
            </CardFooter>
          </Card>
        </div>
      </div>
    </div>
  );
}
