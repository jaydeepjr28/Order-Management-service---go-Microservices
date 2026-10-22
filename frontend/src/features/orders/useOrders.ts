import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { ordersApi } from '@/lib/api/services';
import { useAuthStore } from '@/features/auth/useAuthStore';
import { Order } from '@/types';
import { toast } from 'sonner';

export function useCustomerOrders() {
  const customerID = useAuthStore((state) => state.customerID);
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);

  return useQuery({
    queryKey: ['orders', customerID],
    queryFn: () => ordersApi.getCustomerOrders(customerID!),
    enabled: isAuthenticated && !!customerID,
    staleTime: 1000 * 10,
  });
}

export function useOrderDetail(orderID?: string) {
  const customerID = useAuthStore((state) => state.customerID);
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);

  return useQuery({
    queryKey: ['order', customerID, orderID],
    queryFn: () => ordersApi.getOrder(customerID!, orderID!),
    enabled: isAuthenticated && !!customerID && !!orderID,
    // Poll every 3 seconds until fulfillment reaches "ready" or payment link appears
    refetchInterval: (query) => {
      const data = query.state.data as Order | undefined;
      if (!data) return 3000;
      if (data.Status?.toLowerCase() === 'ready') {
        return false; // Terminal state reached, stop polling
      }
      return 3000;
    },
  });
}

export function useCreateOrder() {
  const queryClient = useQueryClient();
  const customerID = useAuthStore((state) => state.customerID);

  return useMutation({
    mutationFn: (items: Array<{ ID: string; Quantity: number }>) => {
      if (!customerID) throw new Error('You must be signed in to place an order.');
      return ordersApi.createOrder(customerID, items);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['orders', customerID] });
      queryClient.invalidateQueries({ queryKey: ['cart'] });
      toast.success('Order placed successfully!');
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to place order. Please try again.');
    },
  });
}

export function useOrderStatus(orderID?: string) {
  const customerID = useAuthStore((state) => state.customerID);
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);

  return useQuery({
    queryKey: ['order-status', customerID, orderID],
    queryFn: () => ordersApi.getOrderStatus(customerID!, orderID!),
    enabled: isAuthenticated && !!customerID && !!orderID,
    refetchInterval: (query) => {
      const data = query.state.data as { status?: string } | undefined;
      if (!data) return 3000;
      if (data.status?.toLowerCase() === 'ready') {
        return false;
      }
      return 3000;
    },
  });
}

export function useVerifyPayment() {
  const queryClient = useQueryClient();
  const customerID = useAuthStore((state) => state.customerID);

  return useMutation({
    mutationFn: (orderID: string) => {
      if (!customerID) throw new Error('Customer ID is missing');
      return ordersApi.verifyPayment(customerID, orderID);
    },
    onSuccess: (_data, orderID) => {
      queryClient.invalidateQueries({ queryKey: ['order', customerID, orderID] });
      queryClient.invalidateQueries({ queryKey: ['orders', customerID] });
      queryClient.invalidateQueries({ queryKey: ['order-status', customerID, orderID] });
    },
  });
}

export function useCancelOrder() {
  const queryClient = useQueryClient();
  const customerID = useAuthStore((state) => state.customerID);

  return useMutation({
    mutationFn: (orderID: string) => {
      if (!customerID) throw new Error('Customer ID is missing');
      return ordersApi.cancelOrder(customerID, orderID);
    },
    onSuccess: (_data, orderID) => {
      queryClient.invalidateQueries({ queryKey: ['order', customerID, orderID] });
      queryClient.invalidateQueries({ queryKey: ['orders', customerID] });
      queryClient.invalidateQueries({ queryKey: ['admin', 'orders'] });
      queryClient.invalidateQueries({ queryKey: ['inventory'] });
      toast.success('Order cancelled. Saga compensation initiated.');
    },
    onError: (err: Error) => {
      toast.error(err.message || 'Failed to cancel order.');
    },
  });
}


