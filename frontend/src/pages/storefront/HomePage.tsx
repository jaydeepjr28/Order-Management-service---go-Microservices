import React from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useInventory } from '@/features/catalog/useInventory';
import { useCart } from '@/features/cart/useCart';
import { useAuthStore } from '@/features/auth/useAuthStore';
import { SITE_CONFIG } from '@/config/site';
import { formatCurrency } from '@/lib/utils';
import { Button } from '@/components/ui/Button';
import { Card } from '@/components/ui/Card';
import { Skeleton } from '@/components/ui/Skeleton';
import {
  ArrowRight,
  ShoppingBag,
  Sparkles,
  CheckCircle2,
  Clock,
  ShieldCheck,
  PackageCheck,
  Flame,
  Radio,
  Package,
} from 'lucide-react';

export function HomePage() {
  const navigate = useNavigate();
  const { products, isLoading, isBackendSynced } = useInventory();
  const { addItem, isAdding } = useCart();
  const { isAuthenticated, isAdmin } = useAuthStore();

  return (
    <div className="space-y-16 sm:space-y-24">
      {/* 1. Hero Section */}
      <section className="relative overflow-hidden pt-8 pb-12 sm:pt-16 sm:pb-24">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="grid grid-cols-1 lg:grid-cols-12 gap-12 lg:gap-8 items-center">
            {/* Copy Column */}
            <div className="lg:col-span-7 space-y-6 text-left">
              <div className="inline-flex items-center gap-2.5 px-3.5 py-1.5 rounded-full bg-primary-light text-primary text-xs font-semibold tracking-wide border border-primary/20 shadow-subtle">
                <Sparkles className="w-3.5 h-3.5" />
                <span>Distributed Order Architecture</span>
                <span className="w-1 h-1 rounded-full bg-primary" />
                <span className="font-mono text-[11px] opacity-90">v2.0</span>
              </div>

              <h1 className="font-editorial text-4xl sm:text-5xl lg:text-6xl font-bold tracking-tight text-foreground leading-[1.08]">
                {SITE_CONFIG.heroHeadline}
              </h1>

              <p className="text-base sm:text-lg text-muted-foreground leading-relaxed max-w-xl font-normal">
                {SITE_CONFIG.heroSubheadline}
              </p>

              <div className="pt-2 flex flex-wrap items-center gap-3">
                <Button
                  onClick={() => navigate('/products')}
                  variant="primary"
                  size="lg"
                  className="shadow-sm hover:shadow-md"
                  rightIcon={<ArrowRight className="w-4 h-4" />}
                >
                  Explore the Menu
                </Button>

                {isAuthenticated ? (
                  <Button
                    onClick={() => navigate('/orders')}
                    variant="outline"
                    size="lg"
                  >
                    View Your Orders
                  </Button>
                ) : (
                  <Button
                    onClick={() => navigate('/register')}
                    variant="secondary"
                    size="lg"
                  >
                    Create Account
                  </Button>
                )}
              </div>

              {/* Guarantees */}
              <div className="pt-8 grid grid-cols-3 gap-4 border-t border-border/80 text-xs text-muted-foreground">
                <div className="flex items-center gap-2">
                  <CheckCircle2 className="w-4 h-4 text-emerald-600 shrink-0" />
                  <span className="font-medium text-foreground">Real-time Stock</span>
                </div>
                <div className="flex items-center gap-2">
                  <Clock className="w-4 h-4 text-emerald-600 shrink-0" />
                  <span className="font-medium text-foreground">Live Stock Guarantee</span>
                </div>
                <div className="flex items-center gap-2">
                  <ShieldCheck className="w-4 h-4 text-emerald-600 shrink-0" />
                  <span className="font-medium text-foreground">Verified Payments</span>
                </div>
              </div>
            </div>

            {/* Editorial Image Column */}
            <div className="lg:col-span-5 relative">
              <div className="relative rounded-2xl overflow-hidden shadow-elevated border border-border aspect-[4/3] sm:aspect-square bg-muted group">
                <img
                  src="https://images.unsplash.com/photo-1550547660-d9450f859349?auto=format&fit=crop&w=1200&q=85"
                  alt="Artisanal Burger Preparation"
                  className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-700 ease-out"
                />

                {/* Floating Culinary Highlight Tag */}
                <div className="absolute top-4 right-4 bg-card/90 backdrop-blur-md text-foreground text-xs font-semibold px-3 py-1.5 rounded-full shadow border border-border/60 flex items-center gap-1.5">
                  <Flame className="w-3.5 h-3.5 text-amber-500" />
                  <span>Handcrafted Daily</span>
                </div>

                <div className="absolute inset-0 bg-gradient-to-t from-black/75 via-black/20 to-transparent flex flex-col justify-end p-6 sm:p-8 text-white">
                  <span className="text-xs font-semibold tracking-wider uppercase text-emerald-300">
                    Featured Masterpiece
                  </span>
                  <p className="text-2xl font-bold font-editorial mt-1">Artisanal Brioche Smashed Burger</p>
                  <p className="text-xs text-white/85 mt-1 font-mono">
                    ₹199 · Cave-aged cheddar, house brioche, caramelized shallots
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* 2. Real Inventory Highlights */}
      <section className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex flex-col sm:flex-row sm:items-end justify-between mb-8 pb-4 border-b border-border gap-4">
          <div>
            <div className="flex items-center gap-2">
              <span className="text-xs font-bold tracking-widest text-primary uppercase">
                Curated Menu
              </span>
              {!isBackendSynced && (
                <span className="inline-flex items-center gap-1 text-[10px] font-mono font-medium px-2 py-0.5 rounded-full bg-surface border border-border text-muted-foreground">
                  <Radio className="w-2.5 h-2.5 text-amber-500 animate-pulse" />
                  Preview Catalog
                </span>
              )}
            </div>
            <h2 className="text-2xl sm:text-3xl font-bold tracking-tight text-foreground mt-1 font-editorial">
              Fresh from our Culinary Brigade
            </h2>
          </div>
          <Link
            to="/products"
            className="text-sm font-semibold text-primary hover:text-primary-hover inline-flex items-center gap-1 group transition-colors"
          >
            <span>View complete menu</span>
            <ArrowRight className="w-4 h-4 group-hover:translate-x-0.5 transition-transform" />
          </Link>
        </div>

        {/* Product Cards Grid */}
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6 sm:gap-8">
          {isLoading ? (
            Array.from({ length: 3 }).map((_, i) => (
              <Card key={i} className="overflow-hidden p-0 border border-border">
                <Skeleton className="h-56 w-full rounded-none" />
                <div className="p-5 space-y-3">
                  <Skeleton className="h-5 w-2/3" />
                  <Skeleton className="h-4 w-full" />
                  <div className="flex justify-between pt-2">
                    <Skeleton className="h-6 w-16" />
                    <Skeleton className="h-9 w-24" />
                  </div>
                </div>
              </Card>
            ))
          ) : products.length > 0 ? (
            products.slice(0, 6).map((product) => (
              <Card
                key={product.id}
                className="group overflow-hidden border border-border/80 hover:border-primary/40 hover:shadow-elevated transition-all duration-300 flex flex-col justify-between rounded-xl bg-card"
              >
                <div>
                  {/* Product Image */}
                  <div className="relative h-60 overflow-hidden bg-muted">
                    <img
                      src={product.image}
                      alt={product.name}
                      className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-700 ease-out"
                    />
                    {product.badge && (
                      <span className="absolute top-3.5 left-3.5 bg-card/90 backdrop-blur-md text-foreground text-[11px] font-bold px-2.5 py-1 rounded-md shadow-sm border border-border/60">
                        {product.badge}
                      </span>
                    )}
                    <span
                      className={`absolute top-3.5 right-3.5 text-[11px] font-bold px-2.5 py-1 rounded-full shadow-sm flex items-center gap-1.5 ${
                        product.inStock
                          ? 'bg-emerald-600/90 text-white backdrop-blur-sm'
                          : 'bg-red-600/90 text-white backdrop-blur-sm'
                      }`}
                    >
                      <span className="w-1.5 h-1.5 rounded-full bg-white animate-pulse" />
                      {product.inStock ? `${product.stockQuantity} in stock` : 'Out of stock'}
                    </span>
                  </div>

                  {/* Product Details */}
                  <div className="p-5 space-y-2">
                    <div className="flex items-center justify-between">
                      <span className="text-[10px] font-bold tracking-wider uppercase text-primary">
                        {product.category}
                      </span>
                      <span className="text-base font-extrabold text-foreground font-mono">
                        {formatCurrency(product.price)}
                      </span>
                    </div>

                    <Link
                      to={`/products/${product.id}`}
                      className="text-lg font-bold text-foreground hover:text-primary transition-colors block leading-snug"
                    >
                      {product.name}
                    </Link>

                    <p className="text-xs text-muted-foreground line-clamp-2 leading-relaxed">
                      {product.description}
                    </p>
                  </div>
                </div>

                {/* Card Action */}
                <div className="p-5 pt-0">
                  {isAdmin ? (
                    <Button
                      onClick={() => navigate('/admin/inventory')}
                      variant="outline"
                      size="md"
                      className="w-full border-primary/30 text-primary hover:bg-primary/5"
                      leftIcon={<Package className="w-4 h-4" />}
                    >
                      Manage Inventory
                    </Button>
                  ) : (
                    <Button
                      onClick={() => addItem({ productID: product.id, quantity: 1 })}
                      disabled={!product.inStock || isAdding}
                      variant={product.inStock ? 'primary' : 'secondary'}
                      size="md"
                      className="w-full"
                      leftIcon={<ShoppingBag className="w-4 h-4" />}
                    >
                      {product.inStock ? 'Add to Cart' : 'Unavailable'}
                    </Button>
                  )}
                </div>
              </Card>
            ))
          ) : null}
        </div>
      </section>

      {/* 3. Ordering Process Overview */}
      <section className="bg-surface/60 border-y border-border py-16 sm:py-20">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center max-w-2xl mx-auto mb-12">
            <span className="text-xs font-bold tracking-widest text-primary uppercase">
              Our Process
            </span>
            <h2 className="text-2xl sm:text-3xl font-bold tracking-tight text-foreground mt-1 font-editorial">
              Crafted with Care, Fulfilled with Precision
            </h2>
            <p className="text-sm text-muted-foreground mt-2 leading-relaxed">
              Every order is prepared with culinary passion and delivered fresh to your table.
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
            <div className="p-6 sm:p-7 rounded-2xl bg-card border border-border shadow-subtle space-y-4 hover:border-primary/30 transition-colors">
              <div className="w-12 h-12 rounded-xl bg-primary-light text-primary flex items-center justify-center font-bold text-lg">
                01
              </div>
              <h3 className="font-bold text-base text-foreground">Curate & Order</h3>
              <p className="text-xs text-muted-foreground leading-relaxed">
                Select your favorite gourmet items from our fresh daily menu with guaranteed real-time availability.
              </p>
            </div>

            <div className="p-6 sm:p-7 rounded-2xl bg-card border border-border shadow-subtle space-y-4 hover:border-primary/30 transition-colors">
              <div className="w-12 h-12 rounded-xl bg-primary-light text-primary flex items-center justify-center font-bold text-lg">
                02
              </div>
              <h3 className="font-bold text-base text-foreground">Instant Payment</h3>
              <p className="text-xs text-muted-foreground leading-relaxed">
                Enjoy seamless and secure checkout with encrypted payments and instant order confirmation.
              </p>
            </div>

            <div className="p-6 sm:p-7 rounded-2xl bg-card border border-border shadow-subtle space-y-4 hover:border-primary/30 transition-colors">
              <div className="w-12 h-12 rounded-xl bg-primary-light text-primary flex items-center justify-center font-bold text-lg">
                03
              </div>
              <h3 className="font-bold text-base text-foreground">Fresh Preparation</h3>
              <p className="text-xs text-muted-foreground leading-relaxed">
                Our kitchen brigade prepares your meal to perfection and notifies you the moment it is ready.
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* 4. Order Tracking Banner */}
      <section className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pb-8">
        <div className="rounded-2xl bg-primary text-primary-foreground p-8 sm:p-12 shadow-elevated flex flex-col md:flex-row items-center justify-between gap-6">
          <div className="space-y-2 text-center md:text-left">
            <h3 className="text-2xl sm:text-3xl font-bold font-editorial">
              Already placed an order?
            </h3>
            <p className="text-sm text-primary-foreground/80 max-w-lg leading-relaxed">
              Follow your order in real time as our kitchen prepares your meal and updates fulfillment states across the cluster.
            </p>
          </div>
          <Button
            onClick={() => navigate(isAuthenticated ? '/orders' : '/login')}
            variant="secondary"
            size="lg"
            className="shrink-0 font-semibold"
            rightIcon={<PackageCheck className="w-4 h-4" />}
          >
            {isAuthenticated ? 'Track Active Orders' : 'Sign In to Track'}
          </Button>
        </div>
      </section>
    </div>
  );
}
