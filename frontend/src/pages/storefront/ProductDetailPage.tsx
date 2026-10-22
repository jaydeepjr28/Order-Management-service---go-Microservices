import React, { useState } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { useInventory } from '@/features/catalog/useInventory';
import { useCart } from '@/features/cart/useCart';
import { useAuthStore } from '@/features/auth/useAuthStore';
import { formatCurrency } from '@/lib/utils';
import { Button } from '@/components/ui/Button';
import { Card } from '@/components/ui/Card';
import { Skeleton } from '@/components/ui/Skeleton';
import { EmptyState } from '@/components/ui/EmptyState';
import {
  ArrowLeft,
  ShoppingBag,
  Plus,
  Minus,
  CheckCircle2,
  Clock,
  ShieldCheck,
  AlertCircle,
  Package,
} from 'lucide-react';

export function ProductDetailPage() {
  const { productId } = useParams<{ productId: string }>();
  const navigate = useNavigate();
  const { isAdmin } = useAuthStore();
  const { products, isLoading } = useInventory();
  const { addItem, isAdding } = useCart();
  const [quantity, setQuantity] = useState(1);

  const product = products.find((p) => p.id === productId);

  if (isLoading) {
    return (
      <div className="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
        <Skeleton className="h-6 w-32 mb-8" />
        <div className="grid grid-cols-1 md:grid-cols-2 gap-12">
          <Skeleton className="h-96 w-full rounded-2xl" />
          <div className="space-y-4">
            <Skeleton className="h-8 w-3/4" />
            <Skeleton className="h-6 w-1/4" />
            <Skeleton className="h-24 w-full" />
            <Skeleton className="h-12 w-full" />
          </div>
        </div>
      </div>
    );
  }

  if (!product) {
    return (
      <div className="max-w-4xl mx-auto px-4 py-16">
        <EmptyState
          icon={<AlertCircle className="w-6 h-6" />}
          title="Product Not Found"
          description="The requested item does not exist or may have been retired from the active menu."
          actionLabel="Back to Catalog"
          onAction={() => navigate('/products')}
        />
      </div>
    );
  }

  const handleIncrement = () => {
    if (quantity < product.stockQuantity) {
      setQuantity((q) => q + 1);
    }
  };

  const handleDecrement = () => {
    if (quantity > 1) {
      setQuantity((q) => q - 1);
    }
  };

  const handleAddToCart = () => {
    addItem({ productID: product.id, quantity });
  };

  return (
    <div className="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-8">
      {/* Breadcrumb / Back Link */}
      <Link
        to="/products"
        className="inline-flex items-center gap-1.5 text-xs font-semibold text-muted-foreground hover:text-foreground transition-colors"
      >
        <ArrowLeft className="w-4 h-4" />
        <span>Back to Catalog</span>
      </Link>

      <div className="grid grid-cols-1 md:grid-cols-12 gap-8 lg:gap-12 items-start">
        {/* Product Image */}
        <div className="md:col-span-6">
          <div className="rounded-2xl overflow-hidden border border-border bg-card shadow-elevated aspect-square relative">
            <img
              src={product.image}
              alt={product.name}
              className="w-full h-full object-cover"
            />
            {product.badge && (
              <span className="absolute top-4 left-4 bg-card/90 backdrop-blur-sm text-foreground text-xs font-bold px-3 py-1 rounded-md shadow border border-border/60">
                {product.badge}
              </span>
            )}
          </div>
        </div>

        {/* Product Info & Purchase Actions */}
        <div className="md:col-span-6 space-y-6 text-left">
          <div className="space-y-2">
            <span className="text-xs font-bold tracking-widest text-primary uppercase">
              {product.category}
            </span>
            <h1 className="text-3xl sm:text-4xl font-bold font-editorial text-foreground">
              {product.name}
            </h1>
            <div className="flex items-center gap-3 pt-1">
              <span className="text-2xl font-extrabold text-foreground font-mono">
                {formatCurrency(product.price)}
              </span>
              <span
                className={`text-xs font-bold px-2.5 py-0.5 rounded-full ${
                  product.inStock
                    ? 'bg-emerald-500/10 text-emerald-700 dark:text-emerald-400 border border-emerald-500/20'
                    : 'bg-red-500/10 text-red-700 dark:text-red-400 border border-red-500/20'
                }`}
              >
                {product.inStock ? `${product.stockQuantity} in stock` : 'Out of stock'}
              </span>
            </div>
          </div>

          <p className="text-sm text-muted-foreground leading-relaxed">
            {product.description}
          </p>

          {/* Quantity Selector & Add Button */}
          {/* Admin Management vs Customer Purchase Actions */}
          {isAdmin ? (
            <div className="space-y-4 pt-4 border-t border-border">
              <div className="p-4 rounded-xl bg-surface border border-border space-y-2">
                <div className="flex items-center gap-2 text-xs font-semibold text-foreground">
                  <ShieldCheck className="w-4 h-4 text-primary" />
                  <span>Administrator Account Active</span>
                </div>
                <p className="text-xs text-muted-foreground leading-relaxed">
                  Administrator accounts are restricted from placing customer orders. Manage this item's inventory, stock limits, and offerings in the operations console.
                </p>
              </div>
              <Button
                onClick={() => navigate('/admin/inventory')}
                variant="primary"
                size="lg"
                className="w-full"
                leftIcon={<Package className="w-5 h-5" />}
              >
                Manage in Admin Inventory
              </Button>
            </div>
          ) : product.inStock ? (
            <div className="space-y-4 pt-4 border-t border-border">
              <div className="flex items-center gap-4">
                <span className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                  Quantity
                </span>
                <div className="flex items-center border border-border rounded-lg bg-surface">
                  <button
                    onClick={handleDecrement}
                    disabled={quantity <= 1}
                    className="p-2 text-muted-foreground hover:text-foreground disabled:opacity-40 transition-colors"
                    aria-label="Decrease quantity"
                  >
                    <Minus className="w-4 h-4" />
                  </button>
                  <span className="w-12 text-center text-sm font-bold font-mono">
                    {quantity}
                  </span>
                  <button
                    onClick={handleIncrement}
                    disabled={quantity >= product.stockQuantity}
                    className="p-2 text-muted-foreground hover:text-foreground disabled:opacity-40 transition-colors"
                    aria-label="Increase quantity"
                  >
                    <Plus className="w-4 h-4" />
                  </button>
                </div>
                <span className="text-xs text-muted-foreground font-mono">
                  Subtotal: {formatCurrency(product.price * quantity)}
                </span>
              </div>

              <Button
                onClick={handleAddToCart}
                disabled={isAdding}
                variant="primary"
                size="lg"
                className="w-full"
                leftIcon={<ShoppingBag className="w-5 h-5" />}
              >
                Add to Cart ({quantity})
              </Button>
            </div>
          ) : (
            <div className="p-4 rounded-xl bg-red-500/10 border border-red-500/20 text-red-700 dark:text-red-400 text-xs">
              This product is temporarily depleted from inventory. Check back after kitchen restock.
            </div>
          )}

          {/* Feature Badges */}
          <div className="pt-6 border-t border-border space-y-2 text-xs text-muted-foreground">
            <div className="flex items-center gap-2.5">
              <CheckCircle2 className="w-4 h-4 text-emerald-600 shrink-0" />
              <span>Instant inventory reservation on checkout</span>
            </div>
            <div className="flex items-center gap-2.5">
              <Clock className="w-4 h-4 text-emerald-600 shrink-0" />
              <span>Freshly prepared upon order receipt</span>
            </div>
            <div className="flex items-center gap-2.5">
              <ShieldCheck className="w-4 h-4 text-emerald-600 shrink-0" />
              <span>Secure encrypted payment processing</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
