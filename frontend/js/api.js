const API_BASE = '';

const FleetifyAPI = {
  get currentUserId() {
    return localStorage.getItem('fleetify_user_id') || '';
  },

  set currentUserId(id) {
    localStorage.setItem('fleetify_user_id', String(id));
  },

  async request(path, options = {}) {
    const headers = Object.assign({ 'Accept': 'application/json' }, options.headers || {});
    if (this.currentUserId && !options.skipAuth) {
      headers['X-User-ID'] = this.currentUserId;
    }
    const config = Object.assign({}, options, { headers });
    const response = await fetch(API_BASE + path, config);
    const contentType = response.headers.get('content-type') || '';
    let data = null;
    if (contentType.includes('application/json')) {
      data = await response.json();
    } else {
      data = await response.text();
    }
    if (!response.ok) {
      const message = data && data.error ? data.error : 'Request failed';
      throw new Error(message);
    }
    return data;
  },

  getUsers() {
    return this.request('/api/users', { skipAuth: true });
  },

  getVehicles() {
    return this.request('/api/vehicles', { skipAuth: true });
  },

  getMasterItems() {
    return this.request('/api/master-items', { skipAuth: true });
  },

  getReports() {
    return this.request('/api/reports');
  },

  getReport(id) {
    return this.request('/api/reports/' + id);
  },

  createReport(formData) {
    const headers = {};
    if (this.currentUserId) {
      headers['X-User-ID'] = this.currentUserId;
    }
    return fetch(API_BASE + '/api/reports', {
      method: 'POST',
      headers: headers,
      body: formData
    }).then(function (response) {
      return response.json().then(function (data) {
        if (!response.ok) {
          throw new Error(data && data.error ? data.error : 'Request failed');
        }
        return data;
      });
    });
  },

  approveReport(id) {
    return this.request('/api/reports/' + id + '/approve', { method: 'PATCH' });
  },

  completeReport(id, formData) {
    return this.request('/api/reports/' + id + '/complete', {
      method: 'PATCH',
      body: formData
    });
  }
};
