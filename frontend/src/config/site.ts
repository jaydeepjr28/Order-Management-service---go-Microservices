export interface ProductMetadata {
  id: string;
  name: string;
  description: string;
  price: number; // in INR
  category: string;
  image: string;
  badge?: string;
  defaultStock?: number;
}

export const SITE_CONFIG = {
  name: 'STRATUM',
  fullName: 'STRATUM Commerce',
  tagline: 'High-performance distributed provisions, made simple.',
  announcement: '✦ STRATUM Ecosystem · Real-time stock, sub-second payments, & kitchen tracking',
  heroHeadline: 'Good food. Great choices. One easy order.',
  heroSubheadline: 'Explore our curated menu, build your cart in real time, and track your order from kitchen to table.',
  supportEmail: 'concierge@stratum.com',
  currency: 'INR',
  currencySymbol: '₹',
  fallbackImage: 'https://images.unsplash.com/photo-1546069901-ba9599a7e63c?auto=format&fit=crop&w=800&q=80',
};

// Known product catalog enrichment for items returned by stock service + curated preview catalog
export const PRODUCT_CATALOG: Record<string, ProductMetadata> = {
  '1': {
    id: '1',
    name: 'Artisanal Cheese Burger',
    description: 'Double smashed grass-fed beef, melted cave-aged cheddar, crisp butter lettuce, house aioli on toasted brioche.',
    price: 199,
    category: 'Mains',
    image: 'https://images.unsplash.com/photo-1568901346375-23c9450c58cd?auto=format&fit=crop&w=1000&q=80',
    badge: 'Chef’s Pick',
    defaultStock: 20,
  },
  '2': {
    id: '2',
    name: 'Hand-Cut Sea Salt Chips',
    description: 'Crisp russet potato crisps kettle-cooked in small batches, lightly dusted with pure hand-harvested sea salt.',
    price: 99,
    category: 'Sides',
    image: 'https://images.unsplash.com/photo-1566478989037-eec170784d0b?auto=format&fit=crop&w=1000&q=80',
    badge: 'Classic',
    defaultStock: 10,
  },
  '3': {
    id: '3',
    name: 'Parmesan Truffle Fries',
    description: 'Golden fries tossed in aromatic white truffle oil, grated 24-month aged parmesan, and cracked peppercorn.',
    price: 149,
    category: 'Sides',
    image: 'https://images.unsplash.com/photo-1573080496219-bb080dd4f877?auto=format&fit=crop&w=1000&q=80',
    badge: 'Popular',
    defaultStock: 15,
  },
  '4': {
    id: '4',
    name: 'Smoked Bacon Brioche',
    description: 'Hardwood smoked thick bacon, pasture-raised egg, melted sharp provolone, and caramelized shallot relish.',
    price: 229,
    category: 'Mains',
    image: 'https://images.unsplash.com/photo-1550547660-d9450f859349?auto=format&fit=crop&w=1000&q=80',
    badge: 'Signature',
    defaultStock: 12,
  },
  '5': {
    id: '5',
    name: 'Sparkling Botanical Tonic',
    description: 'Crisp cold-pressed elderflower, fresh Meyer lemon, and mountain spring water with effervescent carbonation.',
    price: 89,
    category: 'Beverages',
    image: 'https://images.unsplash.com/photo-1513558161293-cdaf765ed2fd?auto=format&fit=crop&w=1000&q=80',
    badge: 'Organic',
    defaultStock: 25,
  },
  '6': {
    id: '6',
    name: 'Salted Caramel Brioche Tart',
    description: 'Velvety dark chocolate ganache, fleur de sel caramel swirl, nestled inside a flaky golden pastry crust.',
    price: 139,
    category: 'Desserts',
    image: 'https://images.unsplash.com/photo-1509440159596-0249088772ff?auto=format&fit=crop&w=1000&q=80',
    badge: 'New',
    defaultStock: 8,
  },
};

export function getProductMeta(id: string, fallbackName?: string): ProductMetadata {
  if (PRODUCT_CATALOG[id]) {
    return PRODUCT_CATALOG[id];
  }
  return {
    id,
    name: fallbackName || `Product #${id}`,
    description: 'Specialty culinary provision handcrafted by our culinary brigade.',
    price: 149,
    category: 'Provisions',
    image: SITE_CONFIG.fallbackImage,
    defaultStock: 10,
  };
}
