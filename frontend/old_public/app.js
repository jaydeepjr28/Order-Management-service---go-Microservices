const API_BASE = '/api';

let state = {
    token: localStorage.getItem('token') || null,
    customerID: localStorage.getItem('customerID') || null,
    cart: []
};

const views = {
    auth: document.getElementById('auth-view'),
    dashboard: document.getElementById('dashboard-view')
};

const forms = {
    login: document.getElementById('login-form'),
    register: document.getElementById('register-form')
};

function init() {
    if (state.token) {
        showView('dashboard');
        loadInventory();
        loadCart();
    } else {
        showView('auth');
    }
}


function showView(viewName) {
    Object.values(views).forEach(view => view.classList.remove('active'));
    views[viewName].classList.add('active');
}

function switchAuthTab(tab) {
    document.querySelectorAll('.auth-tabs .tab').forEach(t => t.classList.remove('active'));
    document.querySelectorAll('.auth-form').forEach(f => f.classList.remove('active'));

    event.target.classList.add('active');
    forms[tab].classList.add('active');
}

function showSection(sectionName) {
    document.querySelectorAll('.nav-btn').forEach(b => {
        if (!b.classList.contains('logout-btn')) b.classList.remove('active');
    });
    document.querySelectorAll('.content-section').forEach(s => s.classList.remove('active'));

    event.target.classList.add('active');
    document.getElementById(`${sectionName}-section`).classList.add('active');

    if (sectionName === 'inventory') loadInventory();
    if (sectionName === 'cart') loadCart();
    if (sectionName === 'orders') loadOrders();
}

function showToast(message, type = 'success') {
    const toast = document.getElementById('toast');
    toast.textContent = message;
    toast.className = `toast show ${type}`;
    setTimeout(() => {
        toast.className = 'toast hidden';
    }, 3000);
}


async function apiCall(endpoint, method = 'GET', body = null) {
    const headers = {
        'Content-Type': 'application/json'
    };
    if (state.token) {
        headers['Authorization'] = `Bearer ${state.token}`;
    }

    try {
        const response = await fetch(`${API_BASE}${endpoint}`, {
            method,
            headers,
            body: body ? JSON.stringify(body) : null
        });

        if (!response.ok) {
            const err = await response.json().catch(() => ({}));
            throw new Error(err.error || err.message || `Error: ${response.status}`);
        }

        if (response.status === 204) return null;

        const text = await response.text();
        if (!text) return null;
        return JSON.parse(text);
    } catch (error) {
        console.error('API Error:', error);
        throw error;
    }
}


async function handleLogin(e) {
    e.preventDefault();
    const email = document.getElementById('login-email').value;
    const password = document.getElementById('login-password').value;
    const errorEl = document.getElementById('login-error');
    errorEl.textContent = '';

    try {
        const res = await apiCall('/users/login', 'POST', { email, password });
        const token = res.Token || res.token;
        if (token) {
            state.token = token;
            state.customerID = res.CustomerID || res.customerID || email;

            localStorage.setItem('token', state.token);
            localStorage.setItem('customerID', state.customerID);

            showToast('Logged in successfully!');
            showView('dashboard');
            loadInventory();
            loadCart();
        }
    } catch (error) {
        errorEl.textContent = error.message;
    }
}

async function handleRegister(e) {
    e.preventDefault();
    const name = document.getElementById('register-name').value;
    const email = document.getElementById('register-email').value;
    const password = document.getElementById('register-password').value;
    const errorEl = document.getElementById('register-error');
    errorEl.textContent = '';

    try {
        await apiCall('/users/register', 'POST', { name, email, password });
        showToast('Registration successful! Please login.');
        switchAuthTab('login');
        document.querySelector('.auth-tabs .tab:first-child').click();
    } catch (error) {
        errorEl.textContent = error.message;
    }
}

function logout() {
    state.token = null;
    state.customerID = null;
    localStorage.removeItem('token');
    localStorage.removeItem('customerID');
    showView('auth');
    showToast('Logged out');
}


async function loadInventory() {
    try {
        const inventory = await apiCall('/inventory');
        const grid = document.getElementById('inventory-grid');
        grid.innerHTML = '';

        if (!inventory || inventory.length === 0) {
            grid.innerHTML = '<p>No products available right now.</p>';
            return;
        }

        inventory.forEach(item => {
            const card = document.createElement('div');
            card.className = 'product-card';
            card.innerHTML = `
                <h3 class="product-title">${item.name || item.Name || item.ID || item.id}</h3>
                <p class="product-desc">${item.description || item.Description || 'No description available.'}</p>
                <div class="product-price">$${(item.price || item.Price || 0).toFixed(2)}</div>
                <button onclick="addToCart('${item.ID || item.id}')" class="btn add-to-cart-btn">Add to Cart</button>
            `;
            grid.appendChild(card);
        });
    } catch (error) {
        showToast(error.message, 'error');
    }
}

async function loadCart() {
    try {
        const cartItems = await apiCall('/cart');
        const list = document.getElementById('cart-items');
        const count = document.getElementById('cart-count');
        list.innerHTML = '';

        state.cart = (cartItems && (cartItems.items || cartItems.Items)) || [];
        count.textContent = state.cart.length;

        if (state.cart.length === 0) {
            list.innerHTML = '<p>Your cart is empty.</p>';
            return;
        }

        state.cart.forEach(item => {
            const el = document.createElement('div');
            el.className = 'cart-item';
            el.innerHTML = `
                <div class="cart-item-info">
                    <h3>${item.productID || item.ProductID || item.name || 'Product'}</h3>
                    <p>Quantity: ${item.quantity || item.Quantity}</p>
                </div>
                <div class="cart-item-actions">
                    <button onclick="removeFromCart('${item.productID || item.ProductID}')" class="remove-btn">Remove</button>
                </div>
            `;
            list.appendChild(el);
        });
    } catch (error) {
        showToast(error.message, 'error');
    }
}

async function addToCart(productID) {
    try {
        await apiCall('/cart/items', 'POST', { productID, quantity: 1 });
        showToast('Added to cart!');
        loadCart();
    } catch (error) {
        showToast(error.message, 'error');
    }
}

async function removeFromCart(productID) {
    try {
        await apiCall(`/cart/items/${productID}`, 'DELETE');
        showToast('Removed from cart!');
        loadCart();
    } catch (error) {
        showToast(error.message, 'error');
    }
}

async function checkout() {
    if (state.cart.length === 0) {
        showToast('Cart is empty!', 'error');
        return;
    }

    try {
        const orderData = state.cart.map(item => ({
            id: item.productID || item.ProductID,
            quantity: item.quantity || item.Quantity
        }));
        const res = await apiCall(`/customers/${state.customerID}/orders`, 'POST', orderData);

        await apiCall('/cart', 'DELETE').catch(() => console.log("Cart clear error ignored"));

        const paymentLink = res && res.order && (res.order.paymentLink || res.order.PaymentLink);
        if (paymentLink) {
            window.location.href = paymentLink;
        } else {
            showToast('Order placed successfully!');
            showSection('inventory');
        }
    } catch (error) {
        showToast(error.message, 'error');
    }
}

async function loadOrders() {
    try {
        const orders = await apiCall(`/customers/${state.customerID}/orders`);
        const list = document.getElementById('orders-list');
        list.innerHTML = '';

        if (!orders || orders.length === 0) {
            list.innerHTML = '<p>You have no past orders.</p>';
            return;
        }

        orders.forEach(order => {
            const el = document.createElement('div');
            el.className = 'order-item';
            el.style.border = '1px solid rgba(255,255,255,0.2)';
            el.style.padding = '1rem';
            el.style.borderRadius = '8px';
            el.style.background = 'rgba(255,255,255,0.05)';
            
            let itemsHtml = '';
            if (order.items && order.items.length > 0) {
                itemsHtml = `<ul style="margin-top: 10px; padding-left: 20px;">${order.items.map(i => `<li>${i.name || i.Name || 'Item'} (x${i.quantity || i.Quantity || 1})</li>`).join('')}</ul>`;
            }

            el.innerHTML = `
                <div style="display: flex; justify-content: space-between; align-items: center; border-bottom: 1px solid rgba(255,255,255,0.1); padding-bottom: 10px;">
                    <h3>Order #${order.id}</h3>
                    <span style="padding: 4px 8px; background: rgba(255,255,255,0.1); border-radius: 4px; font-weight: bold; text-transform: uppercase; font-size: 0.8rem;">
                        ${order.status}
                    </span>
                </div>
                ${itemsHtml}
            `;
            list.appendChild(el);
        });
    } catch (error) {
        showToast(error.message, 'error');
    }
}

document.addEventListener('DOMContentLoaded', init);
