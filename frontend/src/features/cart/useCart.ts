import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { cartApi } from '@/lib/api/services';
import { useAuthStore } from '@/features/auth/useAuthStore';
import { getProductMeta } from '@/config/site';
import { CartItem } from '@/types';
import { toast } from 'sonner';

export interface EnrichedCartItem {
  productId: string;
  quantity: number;
  name: string;
  price: number;
  image: string;
  totalPrice: number;
}

export function useCart() {
  const queryClient = useQueryClient();
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);

  const cartQuery = useQuery({
    queryKey: ['cart'],
    queryFn: cartApi.getCart,
    enabled: isAuthenticated,
    staleTime: 1000 * 30, // 30 seconds
  });

  const addItemMutation = useMutation({
    mutationFn: async (item: { productID: string; quantity: number }) => {
      if (!isAuthenticated) {
        throw new Error('Please sign in to add items to your cart.');
      }
      return cartApi.addItem(item);
    },
    onSuccess: (data) => {
      queryClient.setQueryData(['cart'], data);
      toast.success('Item added to cart');
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to add item to cart');
    },
  });

  const removeItemMutation = useMutation({
    mutationFn: async (productId: string) => {
      if (!isAuthenticated) {
        throw new Error('Please sign in to update your cart.');
      }
      return cartApi.removeItem(productId);
    },
    onSuccess: (data) => {
      queryClient.setQueryData(['cart'], data);
      toast.info('Item removed from cart');
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to remove item');
    },
  });

  const clearCartMutation = useMutation({
    mutationFn: async () => {
      if (!isAuthenticated) {
        throw new Error('Please sign in to manage your cart.');
      }
      return cartApi.clearCart();
    },
    onSuccess: (data) => {
      queryClient.setQueryData(['cart'], data);
      toast.info('Cart cleared');
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to clear cart');
    },
  });

  const rawItems: CartItem[] = (cartQuery.data?.Items || []) as CartItem[];

  const enrichedItems: EnrichedCartItem[] = rawItems.map((item: any) => {
    const pId = item.ProductID || item.productId || '';
    const qty = Number(item.Quantity ?? item.quantity ?? 1);
    const meta = getProductMeta(pId);
    return {
      productId: pId,
      quantity: qty,
      name: meta.name,
      price: meta.price,
      image: meta.image,
      totalPrice: meta.price * qty,
    };
  });

  const totalQuantity = rawItems.reduce((acc, curr: any) => acc + Number(curr.Quantity ?? curr.quantity ?? 0), 0);
  const totalAmount = enrichedItems.reduce((acc, curr) => acc + curr.totalPrice, 0);

  return {
    items: enrichedItems,
    rawItems,
    totalQuantity,
    totalAmount,
    isLoading: cartQuery.isLoading,
    isError: cartQuery.isError,
    error: cartQuery.error,
    addItem: addItemMutation.mutate,
    isAdding: addItemMutation.isPending,
    removeItem: removeItemMutation.mutate,
    isRemoving: removeItemMutation.isPending,
    clearCart: clearCartMutation.mutate,
    isClearing: clearCartMutation.isPending,
    refetch: cartQuery.refetch,
  };
}
