import React, { useState, useMemo } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useInventory, EnrichedProduct } from '@/features/catalog/useInventory';
import { useCart } from '@/features/cart/useCart';
import { useAuthStore } from '@/features/auth/useAuthStore';
import { formatCurrency } from '@/lib/utils';
import { Button } from '@/components/ui/Button';
import { Card } from '@/components/ui/Card';
import { Input } from '@/components/ui/Input';
import { Skeleton } from '@/components/ui/Skeleton';
import { EmptyState } from '@/components/ui/EmptyState';
import { Search, ShoppingBag, SlidersHorizontal, Radio, X, Package } from 'lucide-react';

export function ProductsPage() {
  const navigate = useNavigate();
  const { isAdmin } = useAuthStore();
  const { products, isLoading, isBackendSynced } = useInventory();
  const { addItem, isAdding } = useCart();

  const [searchQuery, setSearchQuery] = useState('');
  const [selectedCategory, setSelectedCategory] = useState<string>('All');
  const [onlyInStock, setOnlyInStock] = useState(false);
  const [sortBy, setSortBy] = useState<'featured' | 'price-asc' | 'price-desc' | 'name'>('featured');

  const categories = useMemo(() => {
    const set = new Set<string>();
    products.forEach((p) => {
      if (p.category) set.add(p.category);
    });
    return ['All', ...Array.from(set)];
  }, [products]);

  const filteredProducts = useMemo(() => {
    let list = [...products];

    if (searchQuery.trim()) {
      const q = searchQuery.toLowerCase();
      list = list.filter(
        (p) =>
          p.name.toLowerCase().includes(q) ||
          p.description.toLowerCase().includes(q) ||
          p.category.toLowerCase().includes(q)
      );
    }

    if (selectedCategory !== 'All') {
      list = list.filter((p) => p.category === selectedCategory);
    }

    if (onlyInStock) {
      list = list.filter((p) => p.inStock);
    }

    if (sortBy === 'price-asc') {
      list.sort((a, b) => a.price - b.price);
    } else if (sortBy === 'price-desc') {
      list.sort((a, b) => b.price - a.price);
    } else if (sortBy === 'name') {
      list.sort((a, b) => a.name.localeCompare(b.name));
    }

    return list;
  }, [products, searchQuery, selectedCategory, onlyInStock, sortBy]);

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 sm:py-12 space-y-8 text-left">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-end justify-between pb-4 border-b border-border gap-4">
        <div className="space-y-1.5">
          <div className="flex items-center gap-2">
            <span className="text-xs font-bold tracking-widest text-primary uppercase">
              Culinary Store
            </span>
            {!isBackendSynced && (
              <span className="inline-flex items-center gap-1 text-[10px] font-mono font-medium px-2 py-0.5 rounded-full bg-surface border border-border text-muted-foreground">
                <Radio className="w-2.5 h-2.5 text-amber-500 animate-pulse" />
                Preview Mode
              </span>
            )}
          </div>
          <h1 className="text-3xl sm:text-4xl font-bold font-editorial tracking-tight text-foreground">
            Artisanal Provisions
          </h1>
          <p className="text-sm text-muted-foreground">
            Showing {filteredProducts.length} {filteredProducts.length === 1 ? 'item' : 'items'} freshly available to order.
          </p>
        </div>

        {/* Category Pills */}
        <div className="flex items-center gap-1.5 overflow-x-auto pb-1 sm:pb-0 scrollbar-none">
          {categories.map((cat) => (
            <button
              key={cat}
              onClick={() => setSelectedCategory(cat)}
              className={`px-3.5 py-1.5 rounded-full text-xs font-semibold whitespace-nowrap transition-all ${
                selectedCategory === cat
                  ? 'bg-primary text-primary-foreground shadow-sm'
                  : 'bg-surface text-muted-foreground hover:text-foreground hover:bg-surface-hover'
              }`}
            >
              {cat}
            </button>
          ))}
        </div>
      </div>

      {/* Filter and Search Bar */}
      <div className="flex flex-col sm:flex-row gap-4 items-stretch sm:items-center justify-between bg-card p-4 rounded-xl border border-border shadow-subtle">
        <div className="flex-1 max-w-md relative">
          <Input
            placeholder="Search provisions by name or ingredients..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            leftIcon={<Search className="w-4 h-4" />}
            className="w-full pr-8"
          />
          {searchQuery && (
            <button
              onClick={() => setSearchQuery('')}
              className="absolute right-2.5 top-1/2 -translate-y-1/2 p-1 text-muted-foreground hover:text-foreground"
            >
              <X className="w-3.5 h-3.5" />
            </button>
          )}
        </div>

        <div className="flex items-center gap-4">
          {/* Stock Filter Toggle */}
          <label className="flex items-center gap-2 text-xs font-medium text-foreground cursor-pointer select-none">
            <input
              type="checkbox"
              checked={onlyInStock}
              onChange={(e) => setOnlyInStock(e.target.checked)}
              className="rounded border-border text-primary focus:ring-primary w-4 h-4 cursor-pointer"
            />
            <span>In Stock Only</span>
          </label>

          {/* Sort Selector */}
          <div className="flex items-center gap-1.5 pl-3 border-l border-border text-xs">
            <SlidersHorizontal className="w-3.5 h-3.5 text-muted-foreground" />
            <select
              value={sortBy}
              onChange={(e) => setSortBy(e.target.value as any)}
              className="bg-surface border border-border text-foreground text-xs rounded-lg px-2.5 py-2 outline-none focus:border-primary font-medium"
            >
              <option value="featured">Featured</option>
              <option value="price-asc">Price: Low to High</option>
              <option value="price-desc">Price: High to Low</option>
              <option value="name">Name: A to Z</option>
            </select>
          </div>
        </div>
      </div>

      {/* Catalog Grid */}
      {isLoading ? (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6 sm:gap-8">
          {Array.from({ length: 6 }).map((_, i) => (
            <Card key={i} className="overflow-hidden p-0 border border-border">
              <Skeleton className="h-60 w-full rounded-none" />
              <div className="p-5 space-y-3">
                <Skeleton className="h-5 w-2/3" />
                <Skeleton className="h-4 w-full" />
                <div className="flex justify-between pt-2">
                  <Skeleton className="h-6 w-16" />
                  <Skeleton className="h-9 w-24" />
                </div>
              </div>
            </Card>
          ))}
        </div>
      ) : filteredProducts.length === 0 ? (
        <EmptyState
          icon={<Search className="w-6 h-6" />}
          title="No Provisions Found"
          description="We couldn't find any items matching your active search query or category filters."
          actionLabel="Clear Filters"
          onAction={() => {
            setSearchQuery('');
            setSelectedCategory('All');
            setOnlyInStock(false);
            setSortBy('featured');
          }}
        />
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6 sm:gap-8">
          {filteredProducts.map((product: EnrichedProduct) => (
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

                {/* Details */}
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

              {/* Action Button */}
              <div className="p-5 pt-0">
                {isAdmin ? (
                  <Button
                    onClick={() => navigate('/admin/inventory')}
                    variant="outline"
                    size="md"
                    className="w-full font-medium border-primary/30 text-primary hover:bg-primary/5"
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
                    className="w-full font-medium"
                    leftIcon={<ShoppingBag className="w-4 h-4" />}
                  >
                    {product.inStock ? 'Add to Cart' : 'Out of Stock'}
                  </Button>
                )}
              </div>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
