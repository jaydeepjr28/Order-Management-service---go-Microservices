import { useQuery } from '@tanstack/react-query';
import { inventoryApi } from '@/lib/api/services';
import { getProductMeta, PRODUCT_CATALOG, ProductMetadata } from '@/config/site';
import { StockItem } from '@/types';

import { useAuthStore } from '@/features/auth/useAuthStore';

export interface EnrichedProduct extends ProductMetadata {
  stockQuantity: number;
  inStock: boolean;
  priceId?: string;
  isBackendSynced?: boolean;
}

export function useInventory() {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);

  const query = useQuery({
    queryKey: ['inventory'],
    queryFn: inventoryApi.getInventory,
    enabled: isAuthenticated,
    staleTime: 1000 * 30,
    retry: 1,
  });

  const backendItems: StockItem[] = query.data || [];
  const isBackendSynced = backendItems.length > 0;

  let products: EnrichedProduct[] = [];

  if (isBackendSynced) {
    // If backend StockService returned items, prioritize and enrich them
    products = backendItems.map((item: StockItem) => {
      const meta = getProductMeta(item.ID, item.Name);
      const unitPrice = item.Price && item.Price > 0 ? item.Price / 100 : meta.price;
      return {
        ...meta,
        name: item.Name || meta.name,
        price: unitPrice,
        stockQuantity: item.Quantity,
        inStock: item.Quantity > 0,
        priceId: item.PriceID,
        isBackendSynced: true,
      };
    });
  } else {
    // If backend is offline or connecting, show full curated catalog preview
    products = Object.values(PRODUCT_CATALOG).map((meta) => ({
      ...meta,
      stockQuantity: meta.defaultStock || 15,
      inStock: (meta.defaultStock || 15) > 0,
      isBackendSynced: false,
    }));
  }

  return {
    products,
    rawItems: backendItems,
    isBackendSynced,
    isLoading: query.isLoading && !products.length,
    isError: query.isError && isBackendSynced,
    error: query.error,
    refetch: query.refetch,
  };
}
