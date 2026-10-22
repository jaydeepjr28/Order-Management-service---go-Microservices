import React, { useState } from 'react';
import { useInventory, EnrichedProduct } from '@/features/catalog/useInventory';
import {
  useAdminCreateInventory,
  useAdminUpdateInventory,
} from '@/features/admin/useAdmin';
import { formatCurrency } from '@/lib/utils';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { Dialog } from '@/components/ui/Dialog';
import { Input } from '@/components/ui/Input';
import {
  Package,
  Plus,
  Edit2,
  RefreshCw,
  Info,
  CheckCircle,
  AlertTriangle,
} from 'lucide-react';

export function AdminInventoryPage() {
  const { products, isLoading, refetch } = useInventory();
  const createMutation = useAdminCreateInventory();
  const updateMutation = useAdminUpdateInventory();

  // Modals state
  const [isAddOpen, setIsAddOpen] = useState(false);
  const [editingProduct, setEditingProduct] = useState<EnrichedProduct | null>(null);

  // Form states
  const [newId, setNewId] = useState('');
  const [newName, setNewName] = useState('');
  const [newQuantity, setNewQuantity] = useState(10);
  const [newPrice, setNewPrice] = useState(99);

  const [editName, setEditName] = useState('');
  const [editQuantity, setEditQuantity] = useState(0);
  const [editPrice, setEditPrice] = useState(99);

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newId || !newName) return;
    try {
      await createMutation.mutateAsync({
        id: newId,
        name: newName,
        quantity: Number(newQuantity),
        price: Math.round(Number(newPrice) * 100),
      });
      setIsAddOpen(false);
      setNewId('');
      setNewName('');
      setNewQuantity(10);
      setNewPrice(99);
      refetch();
    } catch (err) {
      console.error(err);
    }
  };

  const handleUpdate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!editingProduct) return;
    try {
      await updateMutation.mutateAsync({
        id: editingProduct.id,
        payload: {
          name: editName,
          quantity: Number(editQuantity),
          price: Math.round(Number(editPrice) * 100),
        },
      });
      setEditingProduct(null);
      refetch();
    } catch (err) {
      console.error(err);
    }
  };

  const startEdit = (p: EnrichedProduct) => {
    setEditingProduct(p);
    setEditName(p.name);
    setEditQuantity(p.stockQuantity);
    setEditPrice(p.price || 99);
  };

  return (
    <div className="space-y-6 text-left">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between pb-4 border-b border-border gap-4">
        <div>
          <h1 className="text-2xl sm:text-3xl font-bold tracking-tight text-foreground">
            Inventory Management
          </h1>
          <p className="text-xs sm:text-sm text-muted-foreground mt-0.5">
            Manage catalog items, monitor available stock, and update inventory in real time.
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button
            onClick={() => refetch()}
            variant="outline"
            size="sm"
            leftIcon={<RefreshCw className="w-3.5 h-3.5" />}
          >
            Refresh
          </Button>
          <Button
            onClick={() => setIsAddOpen(true)}
            variant="primary"
            size="sm"
            leftIcon={<Plus className="w-4 h-4" />}
          >
            Add New Item
          </Button>
        </div>
      </div>

      {/* Catalog Availability Note */}
      <div className="p-4 rounded-xl bg-surface border border-border/80 text-xs text-muted-foreground flex items-start gap-3">
        <Info className="w-4 h-4 text-primary shrink-0 mt-0.5" />
        <div className="space-y-1">
          <p className="font-semibold text-foreground">
            Catalog Availability
          </p>
          <p className="leading-relaxed">
            All inventory adjustments are updated live across the customer storefront and reflected immediately during customer checkout.
          </p>
        </div>
      </div>

      {/* Inventory Table */}
      <Card className="border border-border overflow-hidden shadow-subtle">
        <div className="overflow-x-auto">
          <table className="w-full min-w-[700px] text-xs text-left">
            <thead className="bg-surface/70 border-b border-border text-muted-foreground uppercase font-mono tracking-wider">
              <tr>
                <th className="py-3 px-4">Item Details</th>
                <th className="py-3 px-4">SKU / ID</th>
                <th className="py-3 px-4">Catalog Category</th>
                <th className="py-3 px-4">Base Price</th>
                <th className="py-3 px-4">In-Stock Quantity</th>
                <th className="py-3 px-4">Availability</th>
                <th className="py-3 px-4 text-right">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {isLoading ? (
                Array.from({ length: 4 }).map((_, idx) => (
                  <tr key={idx} className="animate-pulse">
                    <td className="py-3 px-4">
                      <div className="flex items-center gap-3">
                        <div className="w-10 h-10 rounded-lg bg-muted shrink-0" />
                        <div className="space-y-1.5 flex-1">
                          <div className="h-4 w-28 bg-muted rounded" />
                          <div className="h-3 w-40 bg-muted rounded" />
                        </div>
                      </div>
                    </td>
                    <td className="py-3 px-4"><div className="h-4 w-10 bg-muted rounded" /></td>
                    <td className="py-3 px-4"><div className="h-4 w-16 bg-muted rounded" /></td>
                    <td className="py-3 px-4"><div className="h-4 w-14 bg-muted rounded" /></td>
                    <td className="py-3 px-4"><div className="h-4 w-8 bg-muted rounded" /></td>
                    <td className="py-3 px-4"><div className="h-5 w-16 bg-muted rounded-full" /></td>
                    <td className="py-3 px-4 text-right"><div className="h-7 w-20 bg-muted rounded ml-auto" /></td>
                  </tr>
                ))
              ) : (
                products.map((product) => (
                <tr key={product.id} className="hover:bg-surface/40 transition-colors">
                  <td className="py-3 px-4">
                    <div className="flex items-center gap-3">
                      <img
                        src={product.image}
                        alt={product.name}
                        className="w-10 h-10 rounded-lg object-cover border border-border"
                      />
                      <div>
                        <p className="font-bold text-foreground text-sm">{product.name}</p>
                        <p className="text-[11px] text-muted-foreground line-clamp-1 max-w-xs">
                          {product.description}
                        </p>
                      </div>
                    </div>
                  </td>
                  <td className="py-3 px-4 font-mono font-semibold text-foreground">
                    #{product.id}
                  </td>
                  <td className="py-3 px-4 font-medium text-muted-foreground">
                    {product.category}
                  </td>
                  <td className="py-3 px-4 font-mono font-bold text-foreground">
                    {formatCurrency(product.price)}
                  </td>
                  <td className="py-3 px-4 font-mono font-bold text-base text-foreground">
                    {product.stockQuantity}
                  </td>
                  <td className="py-3 px-4">
                    <Badge
                      variant={product.inStock ? 'paid' : 'danger'}
                      size="sm"
                    >
                      {product.inStock ? 'In Stock' : 'Depleted'}
                    </Badge>
                  </td>
                  <td className="py-3 px-4 text-right">
                    <Button
                      onClick={() => startEdit(product)}
                      variant="ghost"
                      size="sm"
                      className="h-8 px-2.5 text-primary"
                      leftIcon={<Edit2 className="w-3.5 h-3.5" />}
                    >
                      Edit Product
                    </Button>
                  </td>
                </tr>
              )))}
            </tbody>
          </table>
        </div>
      </Card>

      {/* Add Item Dialog */}
      <Dialog
        isOpen={isAddOpen}
        onClose={() => setIsAddOpen(false)}
        title="Add New Inventory Item"
        description="Add a new product offering to your store catalog with stock and pricing."
      >
        <form onSubmit={handleCreate} className="space-y-4 pt-2">
          <Input
            label="Item Identifier / SKU"
            placeholder="e.g. 4"
            value={newId}
            onChange={(e) => setNewId(e.target.value)}
            required
          />
          <Input
            label="Item Name"
            placeholder="e.g. Truffle Fries"
            value={newName}
            onChange={(e) => setNewName(e.target.value)}
            required
          />
          <div className="grid grid-cols-2 gap-3">
            <Input
              label="Unit Price (₹)"
              type="number"
              min="1"
              step="1"
              placeholder="e.g. 149"
              value={newPrice}
              onChange={(e) => setNewPrice(Number(e.target.value))}
              required
            />
            <Input
              label="Initial Stock Quantity"
              type="number"
              min="0"
              value={newQuantity}
              onChange={(e) => setNewQuantity(Number(e.target.value))}
              required
            />
          </div>

          <div className="flex justify-end gap-2 pt-4 border-t border-border">
            <Button
              type="button"
              onClick={() => setIsAddOpen(false)}
              variant="outline"
              size="sm"
            >
              Cancel
            </Button>
            <Button
              type="submit"
              isLoading={createMutation.isPending}
              variant="primary"
              size="sm"
            >
              Add Item
            </Button>
          </div>
        </form>
      </Dialog>

      {/* Edit Item Dialog */}
      {editingProduct && (
        <Dialog
          isOpen={!!editingProduct}
          onClose={() => setEditingProduct(null)}
          title={`Update SKU #${editingProduct.id}`}
          description="Update product name, price, and available stock quantity."
        >
          <form onSubmit={handleUpdate} className="space-y-4 pt-2">
            <Input
              label="Item Name"
              value={editName}
              onChange={(e) => setEditName(e.target.value)}
              required
            />
            <div className="grid grid-cols-2 gap-3">
              <Input
                label="Unit Price (₹)"
                type="number"
                min="1"
                step="1"
                value={editPrice}
                onChange={(e) => setEditPrice(Number(e.target.value))}
                required
              />
              <Input
                label="Stock Quantity"
                type="number"
                min="0"
                value={editQuantity}
                onChange={(e) => setEditQuantity(Number(e.target.value))}
                required
              />
            </div>

            <div className="flex justify-end gap-2 pt-4 border-t border-border">
              <Button
                type="button"
                onClick={() => setEditingProduct(null)}
                variant="outline"
                size="sm"
              >
                Cancel
              </Button>
              <Button
                type="submit"
                isLoading={updateMutation.isPending}
                variant="primary"
                size="sm"
              >
                Save Changes
              </Button>
            </div>
          </form>
        </Dialog>
      )}
    </div>
  );
}
