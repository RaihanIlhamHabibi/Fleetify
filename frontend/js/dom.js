const DOM = {
  clear(el) {
    while (el.firstChild) {
      el.removeChild(el.firstChild);
    }
  },

  text(tag, text, className) {
    const el = document.createElement(tag);
    el.textContent = text;
    if (className) el.className = className;
    return el;
  },

  option(value, label) {
    const opt = document.createElement('option');
    opt.value = String(value);
    opt.textContent = label;
    return opt;
  },

  statusBadge(status) {
    const map = {
      PENDING_APPROVAL: 'warning',
      APPROVED: 'info',
      COMPLETED: 'success'
    };
    const cls = 'badge text-bg-' + (map[status] || 'secondary');
    return DOM.text('span', status, cls);
  },

  formatDate(iso) {
    if (!iso) return '-';
    const d = new Date(iso);
    return d.toLocaleString('id-ID');
  },

  formatCurrency(n) {
    return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR' }).format(n);
  },

  showAlert(container, message, type) {
    DOM.clear(container);
    const alert = document.createElement('div');
    alert.className = 'alert alert-' + type;
    alert.setAttribute('role', 'alert');
    alert.textContent = message;
    container.appendChild(alert);
  },

  showToast(message, type) {
    const container = document.getElementById('toastContainer');
    const toastEl = document.createElement('div');
    toastEl.className = 'toast align-items-center text-bg-' + (type || 'primary') + ' border-0 show';
    toastEl.setAttribute('role', 'alert');

    const flex = document.createElement('div');
    flex.className = 'd-flex';
    const body = DOM.text('div', message, 'toast-body me-auto');
    const closeBtn = document.createElement('button');
    closeBtn.type = 'button';
    closeBtn.className = 'btn-close btn-close-white me-2 m-auto';
    closeBtn.setAttribute('data-bs-dismiss', 'toast');
    flex.appendChild(body);
    flex.appendChild(closeBtn);
    toastEl.appendChild(flex);
    container.appendChild(toastEl);
    setTimeout(function () {
      toastEl.remove();
    }, 4000);
  }
};
