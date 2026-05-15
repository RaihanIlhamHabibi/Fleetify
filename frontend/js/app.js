(function () {
  let users = [];
  let vehicles = [];
  let masterItems = [];
  let reports = [];
  let detailModal;

  const els = {
    userSelect: document.getElementById('userSelect'),
    currentUserBadge: document.getElementById('currentUserBadge'),
    reportsTableBody: document.getElementById('reportsTableBody'),
    licensePlateInput: document.getElementById('licensePlateInput'),
    vehicleModelInput: document.getElementById('vehicleModelInput'),
    estimateItemsContainer: document.getElementById('estimateItemsContainer'),
    createReportForm: document.getElementById('createReportForm'),
    createAlert: document.getElementById('createAlert'),
    pendingApprovalList: document.getElementById('pendingApprovalList'),
    approvedReportsList: document.getElementById('approvedReportsList'),
    reportDetailBody: document.getElementById('reportDetailBody'),
    exportCsvBtn: document.getElementById('exportCsvBtn'),
    createTabItem: document.getElementById('create-tab-item'),
    approvalTabItem: document.getElementById('approval-tab-item'),
    completeTabItem: document.getElementById('complete-tab-item')
  };

  document.addEventListener('DOMContentLoaded', init);

  async function init() {
    detailModal = new bootstrap.Modal(document.getElementById('reportDetailModal'));
    await loadMasterData();
    bindEvents();
    if (FleetifyAPI.currentUserId) {
      els.userSelect.value = FleetifyAPI.currentUserId;
    }
    onUserChange();
  }

  async function loadMasterData() {
    try {
      users = await FleetifyAPI.getUsers();
      await reloadSuggestions();
      renderUserSelect();
      addEstimateRow();
    } catch (err) {
      DOM.showToast('Gagal memuat data: ' + err.message, 'danger');
    }
  }

  async function reloadSuggestions() {
    vehicles = await FleetifyAPI.getVehicles();
    masterItems = await FleetifyAPI.getMasterItems();
    refreshVehicleDatalists();
    refreshItemNameDatalists();
    refreshMasterItemSelects();
  }

  function refreshItemNameDatalists() {
    var parts = [];
    var services = [];
    masterItems.forEach(function (item) {
      if (item.type === 'PART') {
        parts.push(item.item_name);
      } else if (item.type === 'SERVICE') {
        services.push(item.item_name);
      }
    });
    fillDatalist('itemNameListPart', parts);
    fillDatalist('itemNameListService', services);
  }

  function refreshVehicleDatalists() {
    var models = [];
    vehicles.forEach(function (v) {
      if (v.model && models.indexOf(v.model) === -1) {
        models.push(v.model);
      }
    });
    fillDatalist('vehicleModelList', models);
  }

  function refreshMasterItemSelects() {
    els.estimateItemsContainer.querySelectorAll('.item-select').forEach(function (select) {
      var current = select.value;
      DOM.clear(select);
      masterItems.forEach(function (item) {
        var label = item.item_name + ' [' + item.type + '] - ' + DOM.formatCurrency(item.price);
        select.appendChild(DOM.option(item.id, label));
      });
      if (current) {
        select.value = current;
      }
    });
  }

  function fillDatalist(listId, values) {
    var list = document.getElementById(listId);
    if (!list) return;
    DOM.clear(list);
    var seen = {};
    values.forEach(function (val) {
      var key = String(val).trim();
      if (!key || seen[key]) return;
      seen[key] = true;
      var opt = document.createElement('option');
      opt.value = key;
      list.appendChild(opt);
    });
  }

  function findMasterItem(name, type) {
    var key = name.trim().toLowerCase();
    return masterItems.find(function (item) {
      return item.type === type && item.item_name.trim().toLowerCase() === key;
    });
  }

  function bindEvents() {
    els.userSelect.addEventListener('change', onUserChange);
    document.getElementById('addEstimateRowBtn').addEventListener('click', addEstimateRow);
    document.getElementById('addManualItemBtn').addEventListener('click', addManualEstimateRow);
    els.createReportForm.addEventListener('submit', onCreateReport);
    els.exportCsvBtn.addEventListener('click', exportReportsCsv);
  }

  function currentUser() {
    const id = Number(FleetifyAPI.currentUserId);
    return users.find(function (u) { return u.id === id; });
  }

  function onUserChange() {
    FleetifyAPI.currentUserId = els.userSelect.value;
    const user = currentUser();
    DOM.clear(els.currentUserBadge);
    if (!user) {
      els.currentUserBadge.appendChild(DOM.text('span', 'Pilih pengguna untuk melanjutkan'));
      return;
    }
    const strong = DOM.text('strong', user.username + ' ');
    const badge = DOM.text('span', '(' + user.role + ')', 'badge text-bg-primary ms-1');
    els.currentUserBadge.appendChild(strong);
    els.currentUserBadge.appendChild(badge);
    updateTabsForRole(user.role);
    refreshReports();
  }

  function updateTabsForRole(role) {
    els.createTabItem.classList.toggle('d-none', role !== 'SA');
    els.completeTabItem.classList.toggle('d-none', role !== 'SA');
    els.approvalTabItem.classList.toggle('d-none', role !== 'APPROVAL');
  }

  function renderUserSelect() {
    DOM.clear(els.userSelect);
    els.userSelect.appendChild(DOM.option('', '-- Pilih User --'));
    users.forEach(function (u) {
      els.userSelect.appendChild(DOM.option(u.id, u.username + ' (' + u.role + ')'));
    });
  }

  function addEstimateRow() {
    const row = document.createElement('div');
    row.className = 'row g-2 align-items-end estimate-row estimate-row-master mb-2';

    const colItem = document.createElement('div');
    colItem.className = 'col-md-7';
    const select = document.createElement('select');
    select.className = 'form-select item-select';
    select.required = true;
    masterItems.forEach(function (item) {
      const label = item.item_name + ' [' + item.type + '] - ' + DOM.formatCurrency(item.price);
      select.appendChild(DOM.option(item.id, label));
    });
    colItem.appendChild(select);

    const colQty = document.createElement('div');
    colQty.className = 'col-md-3';
    const qty = document.createElement('input');
    qty.type = 'number';
    qty.min = '1';
    qty.value = '1';
    qty.className = 'form-control qty-input';
    qty.required = true;
    colQty.appendChild(qty);

    const colBtn = document.createElement('div');
    colBtn.className = 'col-md-2';
    const removeBtn = document.createElement('button');
    removeBtn.type = 'button';
    removeBtn.className = 'btn btn-outline-danger btn-sm w-100';
    removeBtn.textContent = 'Hapus';
    removeBtn.addEventListener('click', function () {
      if (els.estimateItemsContainer.children.length > 1) {
        row.remove();
      }
    });
    colBtn.appendChild(removeBtn);

    row.appendChild(colItem);
    row.appendChild(colQty);
    row.appendChild(colBtn);
    els.estimateItemsContainer.appendChild(row);
  }

  function addManualEstimateRow() {
    const row = document.createElement('div');
    row.className = 'row g-2 align-items-end estimate-row estimate-row-manual mb-2';

    const colName = document.createElement('div');
    colName.className = 'col-md-4';
    const nameInput = document.createElement('input');
    nameInput.type = 'text';
    nameInput.className = 'form-control item-name-input';
    nameInput.placeholder = 'Nama jasa / part';
    nameInput.required = true;
    colName.appendChild(nameInput);

    const colType = document.createElement('div');
    colType.className = 'col-md-2';
    const typeSelect = document.createElement('select');
    typeSelect.className = 'form-select item-type-select';
    typeSelect.required = true;
    typeSelect.appendChild(DOM.option('SERVICE', 'Jasa'));
    typeSelect.appendChild(DOM.option('PART', 'Part'));
    colType.appendChild(typeSelect);

    const colPrice = document.createElement('div');
    colPrice.className = 'col-md-2';
    const priceInput = document.createElement('input');
    priceInput.type = 'number';
    priceInput.min = '1';
    priceInput.className = 'form-control item-price-input';
    priceInput.placeholder = 'Harga';
    priceInput.required = true;
    colPrice.appendChild(priceInput);

    const colQty = document.createElement('div');
    colQty.className = 'col-md-2';
    const qty = document.createElement('input');
    qty.type = 'number';
    qty.min = '1';
    qty.value = '1';
    qty.className = 'form-control qty-input';
    qty.required = true;
    colQty.appendChild(qty);

    const colBtn = document.createElement('div');
    colBtn.className = 'col-md-2';
    const removeBtn = document.createElement('button');
    removeBtn.type = 'button';
    removeBtn.className = 'btn btn-outline-danger btn-sm w-100';
    removeBtn.textContent = 'Hapus';
    removeBtn.addEventListener('click', function () {
      if (els.estimateItemsContainer.children.length > 1) {
        row.remove();
      }
    });
    colBtn.appendChild(removeBtn);

    row.appendChild(colName);
    row.appendChild(colType);
    row.appendChild(colPrice);
    row.appendChild(colQty);
    row.appendChild(colBtn);
    els.estimateItemsContainer.appendChild(row);
    bindManualRowAutocomplete(row);
  }

  function bindManualRowAutocomplete(row) {
    var nameInput = row.querySelector('.item-name-input');
    var typeSelect = row.querySelector('.item-type-select');
    var priceInput = row.querySelector('.item-price-input');

    function updateList() {
      nameInput.setAttribute('list', typeSelect.value === 'PART' ? 'itemNameListPart' : 'itemNameListService');
    }

    function applyMaster() {
      var item = findMasterItem(nameInput.value, typeSelect.value);
      if (item) {
        priceInput.value = item.price;
      }
    }

    updateList();
    typeSelect.addEventListener('change', function () {
      updateList();
      applyMaster();
    });
    nameInput.addEventListener('change', applyMaster);
    nameInput.addEventListener('input', function () {
      if (findMasterItem(nameInput.value, typeSelect.value)) {
        applyMaster();
      }
    });
  }

  async function refreshReports() {
    if (!FleetifyAPI.currentUserId) return;
    try {
      reports = await FleetifyAPI.getReports();
      renderReportsTable();
      renderPendingApproval();
      renderApprovedForComplete();
    } catch (err) {
      DOM.showToast(err.message, 'danger');
    }
  }

  function renderReportsTable() {
    DOM.clear(els.reportsTableBody);
    if (!reports.length) {
      const tr = document.createElement('tr');
      const td = document.createElement('td');
      td.colSpan = 6;
      td.className = 'text-center text-muted';
      td.textContent = 'Belum ada laporan';
      tr.appendChild(td);
      els.reportsTableBody.appendChild(tr);
      return;
    }

    reports.forEach(function (r) {
      const tr = document.createElement('tr');
      tr.appendChild(wrapTd(String(r.id)));
      tr.appendChild(wrapTd(r.sa_name));
      tr.appendChild(wrapTd(r.license_plate));
      const statusTd = document.createElement('td');
      statusTd.appendChild(DOM.statusBadge(r.status));
      tr.appendChild(statusTd);
      tr.appendChild(wrapTd(DOM.formatDate(r.created_at)));

      const actionTd = document.createElement('td');
      const btn = document.createElement('button');
      btn.type = 'button';
      btn.className = 'btn btn-sm btn-outline-secondary';
      btn.textContent = 'Detail';
      btn.addEventListener('click', function () { showReportDetail(r.id); });
      actionTd.appendChild(btn);
      tr.appendChild(actionTd);
      els.reportsTableBody.appendChild(tr);
    });
  }

  function wrapTd(text) {
    const td = document.createElement('td');
    td.textContent = text;
    return td;
  }

  function renderPendingApproval() {
    DOM.clear(els.pendingApprovalList);
    const pending = reports.filter(function (r) { return r.status === 'PENDING_APPROVAL'; });
    if (!pending.length) {
      els.pendingApprovalList.appendChild(DOM.text('p', 'Tidak ada laporan menunggu persetujuan.', 'text-muted'));
      return;
    }
    pending.forEach(function (r) {
      els.pendingApprovalList.appendChild(buildActionCard(r, 'approve'));
    });
  }

  function renderApprovedForComplete() {
    DOM.clear(els.approvedReportsList);
    const approved = reports.filter(function (r) { return r.status === 'APPROVED'; });
    if (!approved.length) {
      els.approvedReportsList.appendChild(DOM.text('p', 'Tidak ada laporan siap diselesaikan.', 'text-muted'));
      return;
    }
    approved.forEach(function (r) {
      els.approvedReportsList.appendChild(buildActionCard(r, 'complete'));
    });
  }

  function buildActionCard(report, action) {
    const card = document.createElement('div');
    card.className = 'card mb-3';
    const body = document.createElement('div');
    body.className = 'card-body';

    const title = DOM.text('h6', '#' + report.id + ' - ' + report.license_plate, 'card-title');
    const meta = DOM.text('p', 'SA: ' + report.sa_name + ' | ' + DOM.formatDate(report.created_at), 'card-text small text-muted');
    body.appendChild(title);
    body.appendChild(meta);

    if (action === 'complete') {
      const fileInput = document.createElement('input');
      fileInput.type = 'file';
      fileInput.accept = 'image/*';
      fileInput.className = 'form-control form-control-sm mb-2 proof-input';
      body.appendChild(fileInput);
    }

    const btn = document.createElement('button');
    btn.type = 'button';
    btn.className = 'btn btn-sm ' + (action === 'approve' ? 'btn-success' : 'btn-primary');
    btn.textContent = action === 'approve' ? 'Setujui (APPROVED)' : 'Selesaikan (COMPLETED)';
    btn.addEventListener('click', function () {
      if (action === 'approve') {
        onApprove(report.id);
      } else {
        onComplete(report.id, card.querySelector('.proof-input'));
      }
    });
    body.appendChild(btn);
    card.appendChild(body);
    return card;
  }

  async function onCreateReport(e) {
    e.preventDefault();
    const user = currentUser();
    if (!user || user.role !== 'SA') {
      DOM.showAlert(els.createAlert, 'Hanya Service Advisor yang dapat membuat laporan.', 'warning');
      return;
    }

    if (!els.licensePlateInput.value.trim()) {
      DOM.showAlert(els.createAlert, 'Nomor polisi wajib diisi.', 'warning');
      return;
    }

    const items = [];
    els.estimateItemsContainer.querySelectorAll('.estimate-row').forEach(function (row) {
      if (row.classList.contains('estimate-row-manual')) {
        var name = row.querySelector('.item-name-input').value.trim();
        var price = Number(row.querySelector('.item-price-input').value);
        var qty = Number(row.querySelector('.qty-input').value);
        if (!name || !price || !qty) return;
        items.push({
          item_name: name,
          type: row.querySelector('.item-type-select').value,
          price: price,
          quantity: qty
        });
      } else {
        var itemId = Number(row.querySelector('.item-select').value);
        var qtyMaster = Number(row.querySelector('.qty-input').value);
        if (!itemId || !qtyMaster) return;
        items.push({
          item_id: itemId,
          quantity: qtyMaster
        });
      }
    });
    console.log('[DEBUG] Items collected:', items);

    if (!document.getElementById('odometerInput').value) {
      DOM.showAlert(els.createAlert, 'Odometer wajib diisi.', 'warning');
      return;
    }
    if (!document.getElementById('complaintInput').value.trim()) {
      DOM.showAlert(els.createAlert, 'Keluhan wajib diisi.', 'warning');
      return;
    }
    if (items.length === 0) {
      DOM.showAlert(els.createAlert, 'Tambahkan minimal 1 part/jasa (dari master atau input manual).', 'warning');
      return;
    }

    const formData = new FormData();
    formData.append('license_plate', els.licensePlateInput.value.trim());
    formData.append('vehicle_model', els.vehicleModelInput.value.trim());
    formData.append('odometer', document.getElementById('odometerInput').value);
    formData.append('complaint', document.getElementById('complaintInput').value);
    const itemsJson = JSON.stringify(items);
    console.log('[DEBUG] Items JSON:', itemsJson);
    formData.append('items', itemsJson);

    const photo = document.getElementById('initialPhotoInput').files[0];
    if (photo) {
      formData.append('initial_photo', photo);
    }

    console.log('[DEBUG] FormData keys:', Array.from(formData.keys()));
    try {
      await FleetifyAPI.createReport(formData);
      DOM.showAlert(els.createAlert, 'Laporan berhasil dibuat dengan status PENDING_APPROVAL.', 'success');
      await reloadSuggestions();
      els.createReportForm.reset();
      DOM.clear(els.estimateItemsContainer);
      addEstimateRow();
      await refreshReports();
    } catch (err) {
      DOM.showAlert(els.createAlert, err.message, 'danger');
    }
  }

  async function onApprove(id) {
    try {
      await FleetifyAPI.approveReport(id);
      DOM.showToast('Laporan #' + id + ' disetujui.', 'success');
      await refreshReports();
    } catch (err) {
      DOM.showToast(err.message, 'danger');
    }
  }

  async function onComplete(id, fileInput) {
    const formData = new FormData();
    if (fileInput && fileInput.files[0]) {
      formData.append('proof_photo', fileInput.files[0]);
    }
    try {
      await FleetifyAPI.completeReport(id, formData);
      DOM.showToast('Laporan #' + id + ' selesai.', 'success');
      await refreshReports();
    } catch (err) {
      DOM.showToast(err.message, 'danger');
    }
  }

  async function showReportDetail(id) {
    try {
      const report = await FleetifyAPI.getReport(id);
      DOM.clear(els.reportDetailBody);

      const list = document.createElement('dl');
      list.className = 'row';
      appendDetail(list, 'ID', report.id);
      appendDetail(list, 'Kendaraan', report.vehicle.license_plate + ' (' + report.vehicle.model + ')');
      appendDetail(list, 'SA', report.creator.username);
      appendDetail(list, 'Odometer', report.odometer + ' km');
      appendDetail(list, 'Keluhan', report.complaint);
      appendDetail(list, 'Status', report.status);
      appendDetail(list, 'Tanggal', DOM.formatDate(report.created_at));
      els.reportDetailBody.appendChild(list);

      const table = document.createElement('table');
      table.className = 'table table-sm';
      const thead = document.createElement('thead');
      const headRow = document.createElement('tr');
      ['Item', 'Qty', 'Harga Snapshot', 'Subtotal'].forEach(function (h) {
        const th = document.createElement('th');
        th.textContent = h;
        headRow.appendChild(th);
      });
      thead.appendChild(headRow);
      table.appendChild(thead);

      const tbody = document.createElement('tbody');
      let total = 0;
      (report.items || []).forEach(function (item) {
        const sub = item.price_snapshot * item.quantity;
        total += sub;
        const tr = document.createElement('tr');
        tr.appendChild(wrapTd(item.master_item.item_name));
        tr.appendChild(wrapTd(String(item.quantity)));
        tr.appendChild(wrapTd(DOM.formatCurrency(item.price_snapshot)));
        tr.appendChild(wrapTd(DOM.formatCurrency(sub)));
        tbody.appendChild(tr);
      });
      table.appendChild(tbody);
      els.reportDetailBody.appendChild(table);
      els.reportDetailBody.appendChild(DOM.text('p', 'Total estimasi: ' + DOM.formatCurrency(total), 'fw-bold'));
      detailModal.show();
    } catch (err) {
      DOM.showToast(err.message, 'danger');
    }
  }

  function appendDetail(dl, label, value) {
    const dt = document.createElement('dt');
    dt.className = 'col-sm-3';
    dt.textContent = label;
    const dd = document.createElement('dd');
    dd.className = 'col-sm-9';
    dd.textContent = String(value);
    dl.appendChild(dt);
    dl.appendChild(dd);
  }

  function exportReportsCsv() {
    if (!reports.length) {
      DOM.showToast('Tidak ada data untuk diekspor.', 'warning');
      return;
    }
    const headers = ['ID', 'Nama SA', 'Nomor Polisi', 'Status', 'Tanggal'];
    const rows = reports.map(function (r) {
      return [r.id, r.sa_name, r.license_plate, r.status, DOM.formatDate(r.created_at)];
    });
    const lines = [headers].concat(rows).map(function (row) {
      return row.map(csvEscape).join(',');
    });
    const blob = new Blob([lines.join('\n')], { type: 'text/csv;charset=utf-8;' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'fleetify_reports.csv';
    a.click();
    URL.revokeObjectURL(url);
  }

  function csvEscape(value) {
    const str = String(value).replace(/"/g, '""');
    return '"' + str + '"';
  }
})();
