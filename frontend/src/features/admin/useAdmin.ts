import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { adminApi, inventoryApi } from '@/lib/api/services';
import { useAuthStore } from '@/features/auth/useAuthStore';
import { toast } from 'sonner';

export function useAdminAnalytics() {
  const isAdmin = useAuthStore((state) => state.isAdmin);

  return useQuery({
    queryKey: ['admin', 'analytics'],
    queryFn: adminApi.getAnalytics,
    enabled: isAdmin,
    staleTime: 1000 * 30,
  });
}

export function useAdminOrders() {
  const isAdmin = useAuthStore((state) => state.isAdmin);

  return useQuery({
    queryKey: ['admin', 'orders'],
    queryFn: adminApi.getOrders,
    enabled: isAdmin,
    staleTime: 1000 * 20,
  });
}

export function useAdminCreateInventory() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: adminApi.createInventory,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['inventory'] });
      toast.success('Inventory item created successfully');
    },
    onError: (err: Error) => {
      toast.error(err.message || 'Failed to create inventory item');
    },
  });
}

export function useAdminUpdateInventory() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: { name: string; quantity: number; price?: number } }) =>
      adminApi.updateInventory(id, payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['inventory'] });
      toast.success('Inventory updated successfully');
    },
    onError: (err: Error) => {
      toast.error(err.message || 'Failed to update inventory');
    },
  });
}

export function useAdminCancelOrder() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (orderID: string) => adminApi.cancelOrder(orderID),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'orders'] });
      queryClient.invalidateQueries({ queryKey: ['admin', 'analytics'] });
      queryClient.invalidateQueries({ queryKey: ['inventory'] });
      toast.success('Admin cancellation Saga initiated. Stock restored & refund triggered.');
    },
    onError: (err: Error) => {
      toast.error(err.message || 'Failed to cancel order.');
    },
  });
}

