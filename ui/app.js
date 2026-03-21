const API_BASE = 'http://localhost:4000/v1';

// State
let currentView = 'dashboard';
const state = {
    vehicles: [],
    customers: [],
    rentals: []
};

// DOM Elements
const viewContainer = document.getElementById('view-container');
const pageTitle = document.getElementById('page-title');
const navBtns = document.querySelectorAll('.nav-btn');
const modalOverlay = document.getElementById('modal-container');
const modalTitle = document.getElementById('modal-title');
const modalBody = document.getElementById('modal-body');
const closeModalBtn = document.getElementById('close-modal');

// Init
document.addEventListener('DOMContentLoaded', () => {
    setupNavigation();
    setupModal();
    loadView(currentView);
});

// Navigation
function setupNavigation() {
    navBtns.forEach(btn => {
        btn.addEventListener('click', (e) => {
            const view = e.currentTarget.dataset.view;
            if (view === currentView) return;
            
            navBtns.forEach(b => b.classList.remove('active'));
            e.currentTarget.classList.add('active');
            
            currentView = view;
            loadView(view);
        });
    });
}

function setupModal() {
    closeModalBtn.addEventListener('click', closeModal);
    modalOverlay.addEventListener('click', (e) => {
        if (e.target === modalOverlay) closeModal();
    });
}

// Data Fetching
async function fetchData(endpoint) {
    try {
        const res = await fetch(`${API_BASE}/${endpoint}`);
        if (!res.ok) throw new Error('Network response was not ok');
        return await res.json();
    } catch (err) {
        console.error('Fetch error:', err);
        return null;
    }
}

async function createData(endpoint, data) {
    try {
        const res = await fetch(`${API_BASE}/${endpoint}`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(data)
        });
        return res.ok;
    } catch (err) {
        console.error('Create error:', err);
        return false;
    }
}

async function deleteData(endpoint, id) {
    try {
        const res = await fetch(`${API_BASE}/${endpoint}/${id}`, {
            method: 'DELETE'
        });
        return res.ok || res.status === 204;
    } catch (err) {
        console.error('Delete error:', err);
        return false;
    }
}

// Views
async function loadView(view) {
    viewContainer.innerHTML = '<div class="loading-spinner"></div>';
    pageTitle.textContent = view.charAt(0).toUpperCase() + view.slice(1);

    switch (view) {
        case 'dashboard':
            await renderDashboard();
            break;
        case 'vehicles':
            await renderVehicles();
            break;
        case 'customers':
            await renderCustomers();
            break;
        case 'rentals':
            await renderRentals();
            break;
        default:
            viewContainer.innerHTML = '<h2>View not found</h2>';
    }
}

async function renderDashboard() {
    // Fetch stats
    const [vehiclesData, customersData, rentalsData] = await Promise.all([
        fetchData('vehicles'),
        fetchData('customers'),
        fetchData('rentals')
    ]);

    const vCount = vehiclesData?.vehicles ? vehiclesData.vehicles.length : (vehiclesData?.pagination?.total || 0);
    const cCount = customersData?.customers ? customersData.customers.length : (customersData?.pagination?.total || 0);
    const rCount = rentalsData?.rentals ? rentalsData.rentals.length : (rentalsData?.pagination?.total || 0);

    viewContainer.innerHTML = `
        <div class="grid">
            <div class="card glass">
                <div class="card-title">Total Vehicles</div>
                <div class="card-value">${vCount}</div>
            </div>
            <div class="card glass">
                <div class="card-title">Active Customers</div>
                <div class="card-value">${cCount}</div>
            </div>
            <div class="card glass">
                <div class="card-title">Total Rentals</div>
                <div class="card-value">${rCount}</div>
            </div>
        </div>
    `;
}

async function renderVehicles() {
    const data = await fetchData('vehicles');
    const vehicles = data?.vehicles || [];
    
    viewContainer.innerHTML = `
        <div class="flex-between mb-6">
            <h2>Vehicle Fleet</h2>
            <button class="action-btn btn-primary" onclick="openCreateVehicleModal()">+ Add Vehicle</button>
        </div>
        <div class="table-container glass">
            <table>
                <thead>
                    <tr>
                        <th>ID</th>
                        <th>Make/Model</th>
                        <th>Year</th>
                        <th>License Plate</th>
                        <th>Status</th>
                        <th>Actions</th>
                    </tr>
                </thead>
                <tbody>
                    ${vehicles.map(v => `
                        <tr>
                            <td>${v.id}</td>
                            <td>${v.make} ${v.model}</td>
                            <td>${v.year}</td>
                            <td>${v.license_plate}</td>
                            <td><span class="status ${v.status}">${v.status || 'Available'}</span></td>
                            <td>
                                <button class="action-btn btn-danger" onclick="handleDelete('vehicles', ${v.id})">Delete</button>
                            </td>
                        </tr>
                    `).join('')}
                    ${vehicles.length === 0 ? '<tr><td colspan="6" style="text-align:center">No vehicles found</td></tr>' : ''}
                </tbody>
            </table>
        </div>
    `;
}

async function renderCustomers() {
    const data = await fetchData('customers');
    const customers = data?.customers || [];
    
    viewContainer.innerHTML = `
        <div class="flex-between mb-6">
            <h2>Customers List</h2>
            <button class="action-btn btn-primary" onclick="openCreateCustomerModal()">+ Add Customer</button>
        </div>
        <div class="table-container glass">
            <table>
                <thead>
                    <tr>
                        <th>ID</th>
                        <th>Name</th>
                        <th>Email</th>
                        <th>Phone</th>
                        <th>Status</th>
                        <th>Actions</th>
                    </tr>
                </thead>
                <tbody>
                    ${customers.map(c => `
                        <tr>
                            <td>${c.id}</td>
                            <td>${c.first_name} ${c.last_name}</td>
                            <td>${c.email}</td>
                            <td>${c.phone}</td>
                            <td>${c.is_active ? 'Active' : 'Inactive'}</td>
                            <td>
                                <button class="action-btn btn-danger" onclick="handleDelete('customers', ${c.id})">Delete</button>
                            </td>
                        </tr>
                    `).join('')}
                    ${customers.length === 0 ? '<tr><td colspan="6" style="text-align:center">No customers found</td></tr>' : ''}
                </tbody>
            </table>
        </div>
    `;
}

async function renderRentals() {
    const data = await fetchData('rentals');
    const rentals = data?.rentals || [];
    
    viewContainer.innerHTML = `
        <div class="flex-between mb-6">
            <h2>Rentals Registry</h2>
            <button class="action-btn btn-primary" onclick="openCreateRentalModal()">+ New Rental</button>
        </div>
        <div class="table-container glass">
            <table>
                <thead>
                    <tr>
                        <th>ID</th>
                        <th>Vehicle ID</th>
                        <th>Customer ID</th>
                        <th>Pickup Date</th>
                        <th>Status</th>
                        <th>Actions</th>
                    </tr>
                </thead>
                <tbody>
                    ${rentals.map(r => `
                        <tr>
                            <td>${r.id}</td>
                            <td>${r.vehicle_id}</td>
                            <td>${r.customer_id}</td>
                            <td>${new Date(r.pickup_datetime).toLocaleDateString()}</td>
                            <td>${r.status}</td>
                            <td>
                                <button class="action-btn btn-danger" onclick="handleDelete('rentals', ${r.id})">Delete</button>
                            </td>
                        </tr>
                    `).join('')}
                    ${rentals.length === 0 ? '<tr><td colspan="6" style="text-align:center">No rentals found</td></tr>' : ''}
                </tbody>
            </table>
        </div>
    `;
}

// Handlers
async function handleDelete(resource, id) {
    if (confirm('Are you sure you want to delete this record?')) {
        const success = await deleteData(resource, id);
        if (success) {
            loadView(currentView); // reload current view
        } else {
            alert('Failed to delete relative record');
        }
    }
}

// Modals
function openModal(title, htmlContent) {
    modalTitle.textContent = title;
    modalBody.innerHTML = htmlContent;
    modalOverlay.classList.remove('hidden');
}

function closeModal() {
    modalOverlay.classList.add('hidden');
}

window.openCreateVehicleModal = () => {
    openModal('Add Vehicle', `
        <form id="create-vehicle-form" onsubmit="submitForm(event, 'vehicles')">
            <!-- Go vehicles fields might include: make, model, year, license_plate... -->
            <div class="form-group">
                <label>Make</label>
                <input type="text" name="make" class="form-control" required>
            </div>
            <div class="form-group">
                <label>Model</label>
                <input type="text" name="model" class="form-control" required>
            </div>
            <div class="form-group">
                <label>Year</label>
                <input type="number" name="year" class="form-control" required min="1900" max="2100">
            </div>
            <div class="form-group">
                <label>License Plate</label>
                <input type="text" name="license_plate" class="form-control" required>
            </div>
            <button type="submit" class="btn-submit">Save Vehicle</button>
        </form>
    `);
};

window.openCreateCustomerModal = () => {
    openModal('Add Customer', `
        <form id="create-customer-form" onsubmit="submitForm(event, 'customers')">
            <div class="form-group">
                <label>First Name</label>
                <input type="text" name="first_name" class="form-control" required>
            </div>
            <div class="form-group">
                <label>Last Name</label>
                <input type="text" name="last_name" class="form-control" required>
            </div>
            <div class="form-group">
                <label>Email</label>
                <input type="email" name="email" class="form-control" required>
            </div>
            <div class="form-group">
                <label>Phone</label>
                <input type="text" name="phone" class="form-control" required>
            </div>
            <button type="submit" class="btn-submit">Save Customer</button>
        </form>
    `);
};

window.openCreateRentalModal = () => {
    openModal('New Rental', `
        <form id="create-rental-form" onsubmit="submitForm(event, 'rentals')">
            <div class="form-group">
                <label>Vehicle ID</label>
                <input type="number" name="vehicle_id" class="form-control" required>
            </div>
            <div class="form-group">
                <label>Customer ID</label>
                <input type="number" name="customer_id" class="form-control" required>
            </div>
            <div class="form-group">
                <label>Status</label>
                <input type="text" name="status" value="Active" class="form-control" required>
            </div>
            <button type="submit" class="btn-submit">Create Rental</button>
        </form>
    `);
};

window.submitForm = async (e, endpoint) => {
    e.preventDefault();
    const formData = new FormData(e.target);
    const data = Object.fromEntries(formData.entries());
    
    // Type casting logic based on endpoint (basic handling)
    if (data.year) data.year = parseInt(data.year, 10);
    if (data.vehicle_id) data.vehicle_id = parseInt(data.vehicle_id, 10);
    if (data.customer_id) data.customer_id = parseInt(data.customer_id, 10);

    const btn = e.target.querySelector('button[type="submit"]');
    const originalText = btn.textContent;
    btn.textContent = 'Saving...';
    btn.disabled = true;

    const success = await createData(endpoint, data);
    
    if (success) {
        closeModal();
        loadView(currentView);
    } else {
        btn.textContent = originalText;
        btn.disabled = false;
        alert('Failed to save data. Check the console for details.');
    }
};
