// Queue Management System JavaScript

// Global variables
let websocket = null;
let reconnectAttempts = 0;
const maxReconnectAttempts = 5;

// Create global Queue Manager object
window.QueueManager = {
    initWebSocket: initWebSocket,
    updateOperatorDashboard: updateOperatorDashboard,
    updateDisplayBoard: updateDisplayBoard,
    updateAdminDashboard: updateAdminDashboard,
    showNotification: showNotification,
    websocket: null
};

// Initialize WebSocket connection
function initWebSocket() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws`;
    
    console.log('Attempting to connect to WebSocket:', wsUrl);
    
    try {
        websocket = new WebSocket(wsUrl);
        
        // Make websocket globally accessible
        window.websocket = websocket;
        
        websocket.onopen = function(event) {
            console.log('WebSocket connected successfully');
            reconnectAttempts = 0;
            updateConnectionStatus(true);
            showNotification('Real-time updates connected', 'success');
            
            // Update QueueManager reference
            window.QueueManager.websocket = websocket;
        };
        
        websocket.onmessage = function(event) {
            try {
                console.log('Received WebSocket message:', event.data);
                const data = JSON.parse(event.data);
                updateQueueDisplay(data);
            } catch (error) {
                console.error('Error parsing WebSocket message:', error, 'Raw data:', event.data);
            }
        };
        
        websocket.onclose = function(event) {
            console.log('WebSocket disconnected with code:', event.code, 'reason:', event.reason);
            updateConnectionStatus(false);
            window.websocket = null;
            window.QueueManager.websocket = null;
            attemptReconnect();
        };
        
        websocket.onerror = function(error) {
            console.error('WebSocket error:', error);
            updateConnectionStatus(false);
            showNotification('Connection error occurred', 'warning');
        };
    } catch (error) {
        console.error('Failed to create WebSocket connection:', error);
        updateConnectionStatus(false);
        showNotification('Failed to establish real-time connection', 'error');
    }
}

// Attempt to reconnect WebSocket
function attemptReconnect() {
    if (reconnectAttempts < maxReconnectAttempts) {
        reconnectAttempts++;
        console.log(`Attempting to reconnect... (${reconnectAttempts}/${maxReconnectAttempts})`);
        setTimeout(initWebSocket, 2000 * reconnectAttempts);
    } else {
        console.log('Max reconnection attempts reached');
        showNotification('Connection lost. Please refresh the page.', 'error');
    }
}

// Update connection status indicator
function updateConnectionStatus(connected) {
    const statusElement = document.getElementById('connection-status');
    if (statusElement) {
        statusElement.className = connected ? 'navbar-text connected' : 'navbar-text disconnected';
        statusElement.title = connected ? 'Real-time updates: Connected' : 'Real-time updates: Disconnected';
        
        // Update the icon and text
        const icon = statusElement.querySelector('i');
        const text = statusElement.querySelector('small');
        
        if (icon && text) {
            if (connected) {
                icon.className = 'bi bi-wifi';
                text.textContent = 'Online';
            } else {
                icon.className = 'bi bi-wifi-off';
                text.textContent = 'Offline';
            }
        }
    }
}

// Update queue display with real-time data
function updateQueueDisplay(data) {
    console.log('Received WebSocket data:', data);
    
    // Update service counters
    Object.keys(data).forEach(serviceType => {
        if (serviceType === 'timestamp' || serviceType === 'server_time') return;
        
        const serviceData = data[serviceType];
        if (!serviceData) return;
        
        // Update active count
        const activeElement = document.getElementById(`active-${serviceType.toLowerCase()}`);
        if (activeElement && typeof serviceData.active !== 'undefined') {
            if (activeElement.textContent !== serviceData.active.toString()) {
                activeElement.textContent = serviceData.active;
                animateCounter(activeElement);
            }
        }
        
        // Update processing queue
        const processingElement = document.getElementById(`processing-${serviceType.toLowerCase()}`);
        if (processingElement && serviceData.processing && serviceData.processing.length > 0) {
            const queueNumber = serviceData.processing[0].QueueNumber || 
                               serviceData.processing[0].queue_number || 
                               serviceData.processing[0].Number;
            if (queueNumber) {
                const displayText = `#${queueNumber}`;
                if (processingElement.textContent !== displayText) {
                    processingElement.textContent = displayText;
                    animateElement(processingElement, 'pulse');
                }
            }
        }
        
        // Update waiting counts for operator dashboard
        const waitingElement = document.getElementById(`waiting-${serviceType.toLowerCase()}`);
        if (waitingElement && typeof serviceData.active !== 'undefined') {
            if (waitingElement.textContent !== serviceData.active.toString()) {
                waitingElement.textContent = serviceData.active;
                animateCounter(waitingElement);
            }
        }
        
        // Update total count elements
        const totalElement = document.getElementById(`total-${serviceType.toLowerCase()}`);
        if (totalElement && typeof serviceData.total !== 'undefined') {
            if (totalElement.textContent !== serviceData.total.toString()) {
                totalElement.textContent = serviceData.total;
                animateCounter(totalElement);
            }
        }
        
        // Update completed count elements  
        const completedElement = document.getElementById(`completed-${serviceType.toLowerCase()}`);
        if (completedElement && typeof serviceData.completed !== 'undefined') {
            if (completedElement.textContent !== serviceData.completed.toString()) {
                completedElement.textContent = serviceData.completed;
                animateCounter(completedElement);
            }
        }
    });
    
    // Update display board if on display page
    updateDisplayBoard(data);
    
    // Update operator dashboard if on operator page
    updateOperatorDashboard(data);
    
    // Update queue status if on status page
    updateQueueStatus(data);
    
    // Update admin dashboard if on admin page
    updateAdminDashboard(data);
}

// Update display board
function updateDisplayBoard(data) {
    const displayBoard = document.getElementById('display-board');
    if (!displayBoard) return;
    
    Object.keys(data).forEach(serviceType => {
        if (serviceType === 'timestamp' || serviceType === 'server_time') return;
        
        const serviceData = data[serviceType];
        if (!serviceData) return;
        
        const currentElement = document.getElementById(`current-${serviceType.toLowerCase()}`);
        
        if (currentElement && serviceData.processing && serviceData.processing.length > 0) {
            const currentNumber = serviceData.processing[0].QueueNumber ||
                                 serviceData.processing[0].queue_number ||
                                 serviceData.processing[0].Number;
            
            if (currentNumber && currentElement.textContent !== currentNumber) {
                currentElement.textContent = currentNumber;
                animateElement(currentElement, 'success-bounce');
                playNotificationSound();
            }
        }
        
        // Also update service status divs if they exist
        const serviceStatusDiv = document.getElementById(`service-${serviceType.toLowerCase()}-status`);
        if (serviceStatusDiv) {
            if (serviceData.processing && serviceData.processing.length > 0) {
                const queueNumber = serviceData.processing[0].QueueNumber || 
                                   serviceData.processing[0].queue_number || 
                                   serviceData.processing[0].Number;
                
                serviceStatusDiv.innerHTML = `
                    <h3 class="text-primary">${queueNumber}</h3>
                    <p class="text-muted">Currently Serving</p>
                `;
            } else {
                serviceStatusDiv.innerHTML = `
                    <h4 class="text-muted">No Active Service</h4>
                    <p class="text-muted">Waiting: ${serviceData.active || 0}</p>
                `;
            }
        }
    });
    
    // Also call any page-specific update functions
    if (typeof window.updateServiceStatus === 'function') {
        window.updateServiceStatus(data);
    }
}

// Update operator dashboard
function updateOperatorDashboard(data) {
    const operatorDashboard = document.getElementById('operator-dashboard');
    if (!operatorDashboard) return;
    
    console.log('Updating operator dashboard with data:', data);
    
    // Update waiting counts and other statistics
    Object.keys(data).forEach(serviceType => {
        if (serviceType === 'timestamp' || serviceType === 'server_time') return;
        
        const serviceData = data[serviceType];
        if (!serviceData) return;
        
        // Update waiting count
        const waitingElement = document.getElementById(`waiting-${serviceType}`);
        if (waitingElement && typeof serviceData.active !== 'undefined') {
            if (waitingElement.textContent !== serviceData.active.toString()) {
                waitingElement.textContent = serviceData.active;
                animateCounter(waitingElement);
            }
        }
        
        // Update processing count
        const processingElement = document.getElementById(`processing-${serviceType}`);
        if (processingElement) {
            const processingCount = serviceData.processing ? serviceData.processing.length : 0;
            if (processingElement.textContent !== processingCount.toString()) {
                processingElement.textContent = processingCount;
                animateCounter(processingElement);
            }
        }
        
        // Update completed count
        const completedElement = document.getElementById(`completed-${serviceType}`);
        if (completedElement && typeof serviceData.completed !== 'undefined') {
            if (completedElement.textContent !== serviceData.completed.toString()) {
                completedElement.textContent = serviceData.completed;
                animateCounter(completedElement);
            }
        }
        
        // Update total count
        const totalElement = document.getElementById(`total-${serviceType}`);
        if (totalElement && typeof serviceData.total !== 'undefined') {
            if (totalElement.textContent !== serviceData.total.toString()) {
                totalElement.textContent = serviceData.total;
                animateCounter(totalElement);
            }
        }
        
        // Update current serving information if element exists
        const servingElement = document.getElementById(`serving-${serviceType.toLowerCase()}`);
        if (servingElement) {
            if (serviceData.processing && serviceData.processing.length > 0) {
                const queueNumber = serviceData.processing[0].QueueNumber || 
                                   serviceData.processing[0].queue_number || 
                                   serviceData.processing[0].Number;
                const displayText = queueNumber || 'None';
                if (servingElement.textContent !== displayText) {
                    servingElement.textContent = displayText;
                    if (queueNumber) {
                        animateElement(servingElement, 'success-bounce');
                    }
                }
            } else {
                if (servingElement.textContent !== 'None') {
                    servingElement.textContent = 'None';
                }
            }
        }
    });
    
    // Call the Active Queue update function if it exists
    if (typeof window.updateActiveQueue === 'function') {
        window.updateActiveQueue(data);
    }
}

// Update queue status page
function updateQueueStatus(data) {
    const statusPage = document.getElementById('queue-status');
    if (!statusPage) return;
    
    // Update position and waiting time if available
    // This would need additional server data for current user's queue status
    console.log('Queue status page updated via WebSocket');
}

// Update admin dashboard
function updateAdminDashboard(data) {
    const adminDashboard = document.getElementById('admin-dashboard');
    if (!adminDashboard) return;
    
    // Update service statistics
    Object.keys(data).forEach(serviceType => {
        if (serviceType === 'timestamp' || serviceType === 'server_time') return;
        
        const serviceData = data[serviceType];
        if (!serviceData) return;
        
        // Update stats cards if they exist
        const totalElement = document.querySelector(`[data-service="${serviceType.toLowerCase()}"] .total-count`);
        if (totalElement && typeof serviceData.total !== 'undefined') {
            totalElement.textContent = serviceData.total;
        }
        
        const activeElement = document.querySelector(`[data-service="${serviceType.toLowerCase()}"] .active-count`);
        if (activeElement && typeof serviceData.active !== 'undefined') {
            activeElement.textContent = serviceData.active;
        }
        
        const completedElement = document.querySelector(`[data-service="${serviceType.toLowerCase()}"] .completed-count`);
        if (completedElement && typeof serviceData.completed !== 'undefined') {
            completedElement.textContent = serviceData.completed;
        }
    });
    
    console.log('Admin dashboard updated via WebSocket');
}

// Animate counter changes
function animateCounter(element) {
    element.classList.add('pulse');
    setTimeout(() => {
        element.classList.remove('pulse');
    }, 1000);
}

// Animate element with specified class
function animateElement(element, animationClass) {
    element.classList.add(animationClass);
    setTimeout(() => {
        element.classList.remove(animationClass);
    }, 1000);
}

// Show notification
function showNotification(message, type = 'info') {
    const notification = document.createElement('div');
    notification.className = `alert alert-${type} notification`;
    notification.textContent = message;
    notification.style.cssText = `
        position: fixed;
        top: 20px;
        right: 20px;
        z-index: 9999;
        opacity: 0;
        transition: opacity 0.3s ease;
    `;
    
    document.body.appendChild(notification);
    
    // Fade in
    setTimeout(() => {
        notification.style.opacity = '1';
    }, 100);
    
    // Fade out and remove
    setTimeout(() => {
        notification.style.opacity = '0';
        setTimeout(() => {
            document.body.removeChild(notification);
        }, 300);
    }, 5000);
}

// Play notification sound
function playNotificationSound() {
    try {
        const audio = new Audio('data:audio/wav;base64,UklGRnoGAABXQVZFZm10IBAAAAABAAEAQB8AAEAfAAABAAgAZGF0YQoGAACBhYqFbF1fdJivrJBhNjVgodDbq2EcBj+a2/LDciUFLIHO8tiJNwgZaLvt559NEAxQp+PwtmMcBjiR1/LMeSwFJHfH8N2QQAoUXrTp66hVFApGn+DyvmwhBzuO1PTDdSgEJIDN8t2UQQkXabPt5J5MDAlPqePys2MaB2EIo');
        audio.volume = 0.3;
        audio.play().catch(e => console.log('Audio play failed:', e));
    } catch (error) {
        console.log('Notification sound not available');
    }
}

// Form validation
function validateForm(formId) {
    const form = document.getElementById(formId);
    if (!form) return true;
    
    let isValid = true;
    const inputs = form.querySelectorAll('input[required], select[required]');
    
    inputs.forEach(input => {
        if (!input.value.trim()) {
            isValid = false;
            input.classList.add('is-invalid');
            showFieldError(input, 'This field is required');
        } else {
            input.classList.remove('is-invalid');
            hideFieldError(input);
        }
        
        // Email validation
        if (input.type === 'email' && input.value) {
            const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
            if (!emailRegex.test(input.value)) {
                isValid = false;
                input.classList.add('is-invalid');
                showFieldError(input, 'Please enter a valid email address');
            }
        }
        
        // Phone validation
        if (input.type === 'tel' && input.value) {
            const phoneRegex = /^[\+]?[\d\s\-\(\)]+$/;
            if (!phoneRegex.test(input.value)) {
                isValid = false;
                input.classList.add('is-invalid');
                showFieldError(input, 'Please enter a valid phone number');
            }
        }
    });
    
    return isValid;
}

// Show field error
function showFieldError(input, message) {
    let errorDiv = input.parentNode.querySelector('.invalid-feedback');
    if (!errorDiv) {
        errorDiv = document.createElement('div');
        errorDiv.className = 'invalid-feedback';
        input.parentNode.appendChild(errorDiv);
    }
    errorDiv.textContent = message;
}

// Hide field error
function hideFieldError(input) {
    const errorDiv = input.parentNode.querySelector('.invalid-feedback');
    if (errorDiv) {
        errorDiv.remove();
    }
}

// Copy queue number to clipboard
function copyQueueNumber(number) {
    if (navigator.clipboard) {
        navigator.clipboard.writeText(number).then(() => {
            showNotification(`Queue number ${number} copied to clipboard!`, 'success');
        }).catch(err => {
            console.error('Failed to copy: ', err);
            showNotification('Failed to copy queue number', 'error');
        });
    } else {
        // Fallback for older browsers
        const textArea = document.createElement('textarea');
        textArea.value = number;
        document.body.appendChild(textArea);
        textArea.select();
        try {
            document.execCommand('copy');
            showNotification(`Queue number ${number} copied to clipboard!`, 'success');
        } catch (err) {
            console.error('Fallback copy failed: ', err);
            showNotification('Failed to copy queue number', 'error');
        }
        document.body.removeChild(textArea);
    }
}

// Print queue ticket
function printQueueTicket() {
    window.print();
}

// Confirm action
function confirmAction(message, callback) {
    if (confirm(message)) {
        callback();
    }
}

// Format time
function formatTime(date) {
    return new Date(date).toLocaleTimeString('en-US', {
        hour12: false,
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit'
    });
}

// Calculate waiting time
function calculateWaitingTime(createdAt) {
    const now = new Date();
    const created = new Date(createdAt);
    const diffMs = now - created;
    const diffMins = Math.floor(diffMs / 60000);
    
    if (diffMins < 60) {
        return `${diffMins} min`;
    } else {
        const hours = Math.floor(diffMins / 60);
        const minutes = diffMins % 60;
        return `${hours}h ${minutes}m`;
    }
}

// Initialize page-specific functionality
function initializePage() {
    const path = window.location.pathname;
    console.log('Initializing page for path:', path);
    
    // Initialize WebSocket for real-time updates on pages that need it
    if (path.includes('/display') || 
        path.includes('/operator') || 
        path.includes('/queue/status') ||
        path.includes('/admin') ||
        path === '/') {
        console.log('Starting WebSocket initialization for path:', path);
        initWebSocket();
        
        // Give WebSocket time to initialize before page-specific setup
        setTimeout(() => {
            if (path.includes('/display')) {
                console.log('Setting up display board specific functionality');
                initializeDisplayBoard();
            } else if (path.includes('/operator')) {
                console.log('Setting up operator dashboard specific functionality');
                initializeOperatorDashboard();
            }
        }, 1000);
    } else {
        console.log('WebSocket not needed for path:', path);
    }
    
    // WebSocket-only updates - no auto-refresh
    // Real-time updates handled via WebSocket streaming
    console.log('Page setup complete - using WebSocket streaming for updates');

    
    // Page-specific initializations
    if (path === '/') {
        initializeHomePage();
    } else if (path.includes('/queue/status')) {
        initializeStatusPage();
    } else if (path.includes('/display')) {
        initializeDisplayBoard();
    } else if (path.includes('/operator')) {
        initializeOperatorDashboard();
    } else if (path.includes('/admin')) {
        initializeAdminDashboard();
    }
    
    // WebSocket-only updates - no auto-refresh
    // Real-time updates handled via WebSocket streaming
    console.log('Page setup complete - using WebSocket streaming for updates');
}

// Initialize home page
function initializeHomePage() {
    const form = document.getElementById('join-queue-form');
    if (form) {
        form.addEventListener('submit', function(e) {
            if (!validateForm('join-queue-form')) {
                e.preventDefault();
                return false;
            }
        });
    }
}

// Initialize status page
function initializeStatusPage() {
    // No auto-refresh - using WebSocket streaming for live updates
    
    // Add copy functionality to queue number
    const queueNumber = document.querySelector('.queue-number');
    if (queueNumber) {
        queueNumber.style.cursor = 'pointer';
        queueNumber.title = 'Click to copy';
        queueNumber.addEventListener('click', () => {
            copyQueueNumber(queueNumber.textContent);
        });
    }
}

// Initialize display board
function initializeDisplayBoard() {
    // Full screen functionality
    const fullscreenBtn = document.getElementById('fullscreen-btn');
    if (fullscreenBtn) {
        fullscreenBtn.addEventListener('click', toggleFullscreen);
    }
    
    // No auto-refresh - using WebSocket streaming for live updates
    console.log('Display board initialized with WebSocket streaming only');
}

// Initialize operator dashboard
function initializeOperatorDashboard() {
    // Add confirmation to important actions
    const postponeBtn = document.querySelector('button[formaction*="postpone"]');
    if (postponeBtn) {
        postponeBtn.addEventListener('click', function(e) {
            if (!confirm('Are you sure you want to postpone this customer?')) {
                e.preventDefault();
            }
        });
    }
    
    const completeBtn = document.querySelector('button[formaction*="complete"]');
    if (completeBtn) {
        completeBtn.addEventListener('click', function(e) {
            if (!confirm('Mark this service as completed?')) {
                e.preventDefault();
            }
        });
    }
}

// Initialize admin dashboard
function initializeAdminDashboard() {
    // Add confirmation to user creation
    const createUserForm = document.getElementById('create-user-form');
    if (createUserForm) {
        createUserForm.addEventListener('submit', function(e) {
            if (!validateForm('create-user-form')) {
                e.preventDefault();
                return false;
            }
        });
    }
}

// Toggle fullscreen
function toggleFullscreen() {
    if (!document.fullscreenElement) {
        document.documentElement.requestFullscreen().catch(err => {
            console.error('Error entering fullscreen:', err);
        });
    } else {
        document.exitFullscreen().catch(err => {
            console.error('Error exiting fullscreen:', err);
        });
    }
}

// Keyboard shortcuts
document.addEventListener('keydown', function(e) {
    // F11 for fullscreen on display board
    if (e.key === 'F11' && window.location.pathname.includes('/display')) {
        e.preventDefault();
        toggleFullscreen();
    }
    
    // Ctrl+R to refresh
    if (e.ctrlKey && e.key === 'r') {
        e.preventDefault();
        location.reload();
    }
    
    // Escape to exit fullscreen
    if (e.key === 'Escape' && document.fullscreenElement) {
        document.exitFullscreen();
    }
});

// Initialize when DOM is loaded
document.addEventListener('DOMContentLoaded', initializePage);

// Cleanup when page is unloaded
window.addEventListener('beforeunload', function() {
    if (websocket) {
        websocket.close();
    }
});

// Handle visibility change (tab switching)
document.addEventListener('visibilitychange', function() {
    if (document.visibilityState === 'visible' && websocket && websocket.readyState === WebSocket.CLOSED) {
        initWebSocket();
    }
});

// Export functions for use by other scripts
window.QueueManager = {
    initWebSocket: initWebSocket,
    updateQueueDisplay: updateQueueDisplay,
    updateDisplayBoard: updateDisplayBoard,
    updateOperatorDashboard: updateOperatorDashboard,
    showNotification: showNotification,
    websocket: () => websocket,
    isConnected: () => websocket && websocket.readyState === WebSocket.OPEN
};

// Export functions for use in templates
window.QueueManager = {
    copyQueueNumber,
    printQueueTicket,
    confirmAction,
    showNotification,
    formatTime,
    calculateWaitingTime,
    initWebSocket,
    updateQueueDisplay
};
