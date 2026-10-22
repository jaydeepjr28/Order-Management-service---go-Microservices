import { apiClient } from './client';
import {
  AuthTokens,
  CartResponse,
  CartItem,
  CreateOrderResponse,
  Order,
  OrderStatusResponse,
  StockItem,
  AdminAnalyticsResponse,
} from '@/types';

export const authApi = {
  sendRegistrationOTP: async (payload: { email: string; password: string; name: string }): Promise<{ message: string; email: string; expiresInSeconds: number; cooldownSeconds: number }> => {
    return apiClient<any>('/api/users/register/send-otp', {
      method: 'POST',
      body: JSON.stringify({
        email: payload.email,
        password: payload.password,
        name: payload.name,
      }),
      skipAuth: true,
    });
  },

  verifyRegistrationOTP: async (payload: { email: string; otp: string }): Promise<AuthTokens> => {
    const res = await apiClient<any>('/api/users/register/verify-otp', {
      method: 'POST',
      body: JSON.stringify({
        email: payload.email,
        otp: payload.otp,
      }),
      skipAuth: true,
    });
    return {
      accessToken: res.AccessToken || res.accessToken,
      refreshToken: res.RefreshToken || res.refreshToken,
      customerID: res.CustomerID || res.customerID || payload.email,
    };
  },

  resendRegistrationOTP: async (email: string): Promise<{ message: string; email: string; expiresInSeconds: number; cooldownSeconds: number }> => {
    return apiClient<any>('/api/users/register/resend-otp', {
      method: 'POST',
      body: JSON.stringify({ email }),
      skipAuth: true,
    });
  },

  register: async (payload: { email: string; password: string; name: string }): Promise<AuthTokens> => {
    const res = await apiClient<any>('/api/users/register', {
      method: 'POST',
      body: JSON.stringify({
        Email: payload.email,
        Password: payload.password,
        Name: payload.name,
      }),
      skipAuth: true,
    });
    return {
      accessToken: res.AccessToken || res.accessToken,
      refreshToken: res.RefreshToken || res.refreshToken,
      customerID: res.CustomerID || res.customerID || payload.email,
    };
  },

  login: async (payload: { email: string; password: string }): Promise<AuthTokens> => {
    const res = await apiClient<any>('/api/users/login', {
      method: 'POST',
      body: JSON.stringify({
        Email: payload.email,
        Password: payload.password,
      }),
      skipAuth: true,
    });
    return {
      accessToken: res.AccessToken || res.accessToken,
      refreshToken: res.RefreshToken || res.refreshToken,
      customerID: res.CustomerID || res.customerID || payload.email,
    };
  },

  refreshToken: async (refreshToken: string): Promise<AuthTokens> => {
    const res = await apiClient<any>('/api/users/refresh', {
      method: 'POST',
      body: JSON.stringify({ refreshToken }),
      skipAuth: true,
    });
    return {
      accessToken: res.AccessToken || res.accessToken,
      refreshToken: res.RefreshToken || res.refreshToken || refreshToken,
      customerID: res.CustomerID || res.customerID,
    };
  },
};

export const inventoryApi = {
  getInventory: async (): Promise<StockItem[]> => {
    const res = await apiClient<StockItem[]>('/api/inventory');
    return Array.isArray(res) ? res : [];
  },
};

export const cartApi = {
  getCart: async (): Promise<CartResponse> => {
    const res = await apiClient<any>('/api/cart');
    return {
      CustomerID: res.CustomerID || res.customerID || '',
      Items: res.Items || res.items || [],
    };
  },

  addItem: async (item: { productID: string; quantity: number }): Promise<CartResponse> => {
    const res = await apiClient<any>('/api/cart/items', {
      method: 'POST',
      body: JSON.stringify({
        ProductID: item.productID,
        Quantity: item.quantity,
      }),
    });
    return {
      CustomerID: res.CustomerID || res.customerID || '',
      Items: res.Items || res.items || [],
    };
  },

  removeItem: async (productID: string): Promise<CartResponse> => {
    const res = await apiClient<any>(`/api/cart/items/${encodeURIComponent(productID)}`, {
      method: 'DELETE',
    });
    return {
      CustomerID: res.CustomerID || res.customerID || '',
      Items: res.Items || res.items || [],
    };
  },

  clearCart: async (): Promise<CartResponse> => {
    const res = await apiClient<any>('/api/cart', {
      method: 'DELETE',
    });
    return {
      CustomerID: res.CustomerID || res.customerID || '',
      Items: res.Items || res.items || [],
    };
  },
};

export const ordersApi = {
  createOrder: async (
    customerID: string,
    items: Array<{ ID: string; Quantity: number }>
  ): Promise<CreateOrderResponse> => {
    return apiClient<CreateOrderResponse>(`/api/customers/${encodeURIComponent(customerID)}/orders`, {
      method: 'POST',
      body: JSON.stringify(items),
    });
  },

  getCustomerOrders: async (customerID: string): Promise<Order[]> => {
    const res = await apiClient<Order[]>(`/api/customers/${encodeURIComponent(customerID)}/orders`);
    return Array.isArray(res) ? res : [];
  },

  getOrder: async (customerID: string, orderID: string): Promise<Order> => {
    return apiClient<Order>(`/api/customers/${encodeURIComponent(customerID)}/orders/${encodeURIComponent(orderID)}`);
  },

  getOrderStatus: async (customerID: string, orderID: string): Promise<OrderStatusResponse> => {
    return apiClient<OrderStatusResponse>(`/api/customers/${encodeURIComponent(customerID)}/orders/${encodeURIComponent(orderID)}/status`);
  },

  verifyPayment: async (customerID: string, orderID: string): Promise<{ status: string; verified: boolean; message?: string }> => {
    return apiClient<{ status: string; verified: boolean; message?: string }>(
      `/api/customers/${encodeURIComponent(customerID)}/orders/${encodeURIComponent(orderID)}/verify-payment`,
      {
        method: 'POST',
      }
    );
  },

  cancelOrder: async (customerID: string, orderID: string): Promise<{ message: string; status: string; orderID: string }> => {
    return apiClient<{ message: string; status: string; orderID: string }>(
      `/api/customers/${encodeURIComponent(customerID)}/orders/${encodeURIComponent(orderID)}/cancel`,
      {
        method: 'POST',
      }
    );
  },
};

export const adminApi = {
  getOrders: async (): Promise<any> => {
    return apiClient<any>('/api/admin/orders');
  },

  getAnalytics: async (): Promise<AdminAnalyticsResponse> => {
    return apiClient<AdminAnalyticsResponse>('/api/admin/analytics');
  },

  cancelOrder: async (orderID: string): Promise<{ message: string; status: string; orderID: string }> => {
    return apiClient<{ message: string; status: string; orderID: string }>(
      `/api/admin/orders/${encodeURIComponent(orderID)}/cancel`,
      {
        method: 'POST',
      }
    );
  },

  createInventory: async (payload: { id: string; name: string; quantity: number; price?: number }): Promise<any> => {
    return apiClient<any>('/api/admin/inventory', {
      method: 'POST',
      body: JSON.stringify({
        ID: payload.id,
        Name: payload.name,
        Quantity: payload.quantity,
        Price: payload.price,
      }),
    });
  },

  updateInventory: async (id: string, payload: { name: string; quantity: number; price?: number }): Promise<any> => {
    return apiClient<any>(`/api/admin/inventory/${encodeURIComponent(id)}`, {
      method: 'PUT',
      body: JSON.stringify({
        Name: payload.name,
        Quantity: payload.quantity,
        Price: payload.price,
      }),
    });
  },
};


