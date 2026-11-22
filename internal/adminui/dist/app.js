'use strict';
const backendPrefix = new URL('../', document.currentScript.src).pathname.replace(/\/$/, '');
const $ = (s, root = document) => root.querySelector(s);
const $$ = (s, root = document) => [...root.querySelectorAll(s)];
const esc = v => String(v ?? '').replace(/[&<>"']/g, c => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));
const licenses = ['A1','A','B1','B','C1','C','D1','D2','D','BE','C1E','CE','D1E','D2E','DE'];
const categoryTypes=['UNCATEGORIZED','CONCEPT','RULE','SIGN','SA_HINH','CULTURE','TECHNIQUE','STRUCTURE'];
const signTypes=['PROHIBITION','WARNING','MANDATORY','INFORMATION','SUPPLEMENTARY','OTHER'];
const titles = {signs:'Học biển báo',templates:'Cấu hình đề random',overview:'Tổng quan',questions:'Ngân hàng câu hỏi',categories:'Chủ đề',tests:'Đề thi',import:'Nhập dữ liệu',api:'Danh mục API'};
let session = null;
try { session = JSON.parse(sessionStorage.getItem('driving-admin')); } catch { sessionStorage.removeItem('driving-admin'); }
let view = 'overview', page = 1, categories = [], rows = [], filters = {}, renderVersion = 0, refreshPromise, toastTimer;
let publishSelection = new Set(), publishResult = null, publishing = false;
let editing = null, selected = new Map(), pickerPage = 1;
function remember(value) { session = value; if (value) sessionStorage.setItem('driving-admin', JSON.stringify(value)); else sessionStorage.removeItem('driving-admin'); }
function logout() { publishSelection.clear(); publishResult=null; remember(null); renderVersion++; $('#editor').close(); $('#workspace').hidden = true; $('#login-screen').hidden = false; $('#login-form').reset(); }
function notify(message) { $('#toast').textContent = message; $('#toast').hidden = false; clearTimeout(toastTimer); toastTimer = setTimeout(() => $('#toast').hidden = true, 6000); }
async function request(path, method = 'GET', body, retry = true) {
 const headers = {};
 if (session?.access_token) headers.Authorization = `Bearer ${session.access_token}`;
 if (body !== undefined) headers['Content-Type'] = 'application/json';
 let response;
 try { response = await fetch(backendPrefix + path, {method, headers, body:body === undefined ? undefined : JSON.stringify(body)}); }
 catch { throw new Error('Không kết nối được backend. Kiểm tra server rồi thử lại.'); }
 if (response.status === 401 && retry && session?.refresh_token && !path.startsWith('/auth/')) {
  if (!refreshPromise) {
   const current = session;
   refreshPromise = request('/auth/refresh', 'POST', {refresh_token:current.refresh_token}, false)
    .then(result => { if (session === current) remember({...current,...result.data}); else throw new Error('Phiên đăng nhập đã thay đổi.'); })
    .catch(error => { if (session === current) logout(); throw error; }).finally(() => refreshPromise = null);
  }
  await refreshPromise;
  return request(path, method, body, false);
 }
 const data = response.status === 204 ? null : await response.json().catch(() => ({}));
 if (!response.ok) {
  if (response.status === 401 && !path.startsWith('/auth/')) logout();
  throw new Error(data?.message || `Yêu cầu thất bại (HTTP ${response.status}).`);
 }
 return data;
}
async function busy(button, action, errorTarget) {
 if (button?.disabled) return;
 if (button) button.disabled = true;
 try { await action(); } catch (error) { if (errorTarget) errorTarget.textContent = error.message; else notify(error.message); }
 finally { if (button) button.disabled = button.dataset.unavailable === 'true'; }
}
function options(values, current, empty) { return (empty ? `<option value="">${esc(empty)}</option>` : '') + values.map(v => {const [id,name] = Array.isArray(v) ? v : [v,v]; return `<option value="${esc(id)}" ${String(id) === String(current) ? 'selected' : ''}>${esc(name)}</option>`;}).join(''); }
const catOptions = current => options(categories.map(c => [c.id,c.name]),current);
const badge = published => `<span class="badge ${published?'live':''}">${published?'Đã xuất bản':'Bản nháp'}</span>`;
const heading = (title, desc, actions='') => `<div class="page-heading"><div><span class="eyebrow">CONTENT STUDIO</span><h1>${title}</h1><p>${desc}</p></div><div class="actions">${actions}</div></div>`;
const actionButton = (action,text,extra='') => `<button data-action="${action}" ${extra}>${text}</button>`;
function field(name,label,value='',type='text',attrs='') { return `<label>${label}<input name="${name}" type="${type}" value="${esc(value)}" ${attrs}></label>`; }
function area(name,label,value='',attrs='') {return `<label>${label}<textarea name="${name}" ${attrs}>${esc(value)}</textarea></label>`;}
function check(name,label,value) {return `<label class="check"><input name="${name}" type="checkbox" ${value?'checked':''}>${label}</label>`;}
function pager(total,limit=20) { return `<div class="pager"><span>${total} bản ghi · Trang ${page}/${Math.max(1,Math.ceil(total/limit))}</span><div class="actions">${actionButton('prev','← Trước',page<=1?'disabled':'')}${actionButton('next','Tiếp →',page*limit>=total?'disabled':'')}</div></div>`; }
async function enter() {
 const user = await request('/auth/getProfile');
 if (user.role !== 'admin') { logout(); throw new Error('Tài khoản chưa có quyền admin. Nhờ chủ server cấp quyền bằng CLI.'); }
 categories = (await request('/api/v1/categories')).data || [];
 $('#user-name').textContent = user.name || user.email;
 $('#login-screen').hidden = true; $('#workspace').hidden = false;
 await navigate('overview');
}
$('#login-form').addEventListener('submit', event => {event.preventDefault(); busy($('button[type=submit]',event.target),async () => {
 $('#login-error').textContent = '';
 const data = Object.fromEntries(new FormData(event.target));
 const result = await request('/auth/login','POST',data);
 remember(result.data);
 await enter(); event.target.reset();
}, $('#login-error'));});
$('#logout').onclick = logout;
$$('[data-nav]').forEach(b => b.onclick = () => busy(b,()=>navigate(b.dataset.nav)));
async function navigate(next) {publishSelection.clear(); publishResult=null; view=next; page=1; filters={}; await render();}
async function render() {
 const version = ++renderVersion, current = view;
 $('#breadcrumb').textContent=titles[current];
 $$('[data-nav]').forEach(b => b.classList.toggle('active',b.dataset.nav===current));
 $('#content').innerHTML='<div class="loading">Đang tải dữ liệu…</div>';
 try {
  let html;
  if (current==='overview') html=await overview();
  if (current==='categories') html=await categoryList();
  if (current==='questions'||current==='tests') html=await recordList(current);
  if (current==='signs'||current==='templates')html=await mobileAdminList(current);
  if (current==='import') html=importPage();
  if (current==='api') html=await apiPage();
  if (version!==renderVersion) return;
  $('#content').innerHTML=html;
  if(current==='questions') updatePublishSelection();
 } catch(error) {if(version===renderVersion) $('#content').innerHTML=`<div class="card"><h2>Chưa tải được dữ liệu</h2><p class="error">${esc(error.message)}</p>${actionButton('reload','Thử lại')}</div>`;}
}
async function overview() {
 const [qs,ts,cs] = await Promise.all([request('/api/v1/admin/questions?limit=1'),request('/api/v1/admin/tests?limit=5'),request('/api/v1/categories')]);
 categories=cs.data||[];
 return heading('Nội dung tốt. Khởi đầu vững.','Quản lý học liệu và chuẩn bị những bài thi chất lượng.')+
 `<section class="hero"><div><span class="eyebrow">TỪ NGÂN HÀNG CÂU HỎI ĐẾN BÀI THI</span><h2>Không gian cho người biên soạn</h2><p>Tổ chức câu hỏi theo chủ đề, rà soát đáp án và xuất bản khi nội dung đã sẵn sàng cho người học.</p><div class="actions">${actionButton('new-question','＋ Thêm câu hỏi','class="primary"')}${actionButton('go-tests','Quản lý đề thi →')}</div></div><div class="hero-mark" aria-hidden="true">↗</div></section>`+
 `<div class="stats"><div class="stat"><span>NGÂN HÀNG CÂU HỎI</span><strong>${qs.total}</strong><span>Bản nháp và đã xuất bản</span></div><div class="stat"><span>CHỦ ĐỀ</span><strong>${categories.length}</strong><span>Nhóm nội dung học tập</span></div><div class="stat"><span>ĐỀ THI</span><strong>${ts.total}</strong><span>Đề cố định và được sinh</span></div></div>`+
 `<section class="card"><div class="card-head"><h3>Đề thi mới tạo</h3>${actionButton('go-tests','Xem tất cả →','class="quiet"')}</div>${testTable(ts.data||[])}</section>`+
 `<div class="notice">Quy trình gợi ý: tạo chủ đề → thêm câu hỏi và lời giải → xuất bản câu hỏi → tạo đề → kiểm tra và xuất bản đề.</div>`;
}
async function categoryList() {
 categories=(await request('/api/v1/categories')).data||[]; rows=categories;
 return heading('Chủ đề','Sắp xếp ngân hàng câu hỏi thành các nhóm dễ quản lý.',actionButton('new-category','＋ Thêm chủ đề','class="primary"'))+
 `<div class="table-wrap"><table><thead><tr><th>ID</th><th>Chủ đề / Loại</th><th>Mô tả</th><th>Thao tác</th></tr></thead><tbody>${categories.map(c=>`<tr><td>#${c.id}</td><td><strong>${esc(c.name)}</strong><br><span class="badge">${esc(c.type)}</span></td><td>${esc(c.description||'—')}</td><td>${rowActions(c.id)}</td></tr>`).join('')||'<tr><td colspan="4" class="empty">Chưa có chủ đề. Tạo chủ đề đầu tiên để thêm câu hỏi.</td></tr>'}</tbody></table></div>`;
}
function rowActions(id) {return `<div class="row-actions">${actionButton('edit','Sửa',`data-id="${id}"`)}${actionButton('delete','Xóa',`data-id="${id}" class="danger"`)}</div>`;}
function testTable(data) {return `<div class="table-wrap"><table><thead><tr><th>Đề thi</th><th>Hạng</th><th>Cấu hình</th><th>Trạng thái</th><th>Thao tác</th></tr></thead><tbody>${data.map(t=>`<tr><td class="title"><strong>${esc(t.name)}</strong><small>#${t.id} · ${esc(t.version)}</small></td><td><span class="badge">${esc(t.license_type)}</span></td><td>${t.total_questions} câu · ${t.duration_minutes} phút<br><small>Đạt ${t.passing_score} điểm${t.critical_fail?' · Có điểm liệt':''}</small></td><td>${badge(t.published)}</td><td>${actionButton('edit-test','Mở đề →',`data-id="${t.id}"`)} ${view==='tests'?actionButton('delete','Xóa',`data-id="${t.id}" class="danger"`):''}</td></tr>`).join('')||'<tr><td colspan="5" class="empty">Chưa có đề thi phù hợp.</td></tr>'}</tbody></table></div>`;}
async function recordList(kind) {
 const params=new URLSearchParams({page,limit:20,...filters});
 const data=await request(`/api/v1/admin/${kind}?${params}`); rows=data.data||[];
 const isQ=kind==='questions';
 let html=heading(titles[kind],isQ?'Biên tập nội dung, đáp án và lời giải trước khi xuất bản.':'Chọn câu hỏi hoặc sinh đề theo số lượng từng chủ đề.',isQ?actionButton('new-question','＋ Thêm câu hỏi','class="primary"'):actionButton('generate','Sinh đề')+actionButton('new-test','＋ Tạo đề','class="primary"'));
 html+=`<form id="filters" class="toolbar">${isQ?`<label>Tìm kiếm<input name="search" value="${esc(filters.search||'')}" placeholder="Nội dung hoặc tiêu đề câu hỏi"></label>`:''}<label>Hạng bằng<select name="license_type">${options(licenses,filters.license_type,'Tất cả hạng')}</select></label>${isQ?`<label>Chủ đề<select name="category_id">${options(categories.map(c=>[c.id,c.name]),filters.category_id,'Tất cả chủ đề')}</select></label><label>Điểm liệt<select name="is_critical">${options([['true','Câu điểm liệt'],['false','Câu thường']],filters.is_critical,'Tất cả câu')}</select></label>`:''}<label>Loại chủ đề<select name="category_type">${options(categoryTypes,filters.category_type,'Tất cả loại')}</select></label><button type="submit">Lọc dữ liệu</button></form>`;
 if(isQ) html+=publishToolbar();
 html+=isQ?`<div class="table-wrap"><table><thead><tr><th><input id="select-question-page" type="checkbox" aria-label="Chọn tất cả câu nháp trên trang"></th><th>Câu hỏi</th><th>Phân loại</th><th>Trạng thái</th><th>Thao tác</th></tr></thead><tbody>${rows.map(q=>`<tr><td><input type="checkbox" data-publish-id="${q.id}" aria-label="Chọn câu hỏi ${q.id}" ${q.published?'disabled title="Câu đã xuất bản"':''} ${publishSelection.has(q.id)&&!q.published?'checked':''}></td><td class="title"><strong>${esc(q.title)}</strong><small>#${q.id} · ${esc(q.content.length>125?q.content.slice(0,125)+'…':q.content)}</small></td><td><span class="badge">${esc(q.license_type)}</span> ${q.is_critical?'<span class="badge critical">Điểm liệt</span>':''}<br><small>${esc(q.category?.name||'')}</small></td><td>${badge(q.published)}</td><td>${rowActions(q.id)}</td></tr>`).join('')||'<tr><td colspan="5" class="empty">Không có câu hỏi phù hợp. Thử đổi bộ lọc hoặc thêm câu hỏi mới.</td></tr>'}</tbody></table></div>`:testTable(rows);
 return html+pager(data.total);
}
function publishToolbar() {
 const result=publishResult;
 return `<section class="bulk-publish"><div><strong id="publish-count">Đã chọn ${publishSelection.size} câu</strong><p class="muted">Chọn câu nháp ở một hoặc nhiều trang (tối đa 500). Đổi bộ lọc sẽ bỏ lựa chọn.</p></div><div class="actions">${actionButton('clear-publish-selection','Bỏ chọn')}${actionButton('publish-selected','Xuất bản các câu đã chọn','class="primary"')}</div></section>`+
 (result?`<section class="notice" role="status"><strong>${result.published_ids.length} câu đã xuất bản · ${result.failed.length} câu chưa xuất bản.</strong>${result.failed.length?`<p>Sửa các câu dưới đây rồi thử lại; lựa chọn của các câu lỗi được giữ nguyên.</p><ul>${result.failed.map(item=>`<li>Câu #${item.id}: ${esc(item.message)}</li>`).join('')}</ul>`:''}</section>`:'');
}
function updatePublishSelection() {
 const checkboxes=$$('[data-publish-id]:not(:disabled)');
 $$('[data-publish-id]:disabled').forEach(input=>publishSelection.delete(Number(input.dataset.publishId)));
 for(const input of checkboxes)input.checked=publishSelection.has(Number(input.dataset.publishId));
 const chosen=checkboxes.filter(input=>input.checked).length, all=$('#select-question-page');
 if(all){all.checked=checkboxes.length>0&&chosen===checkboxes.length;all.indeterminate=chosen>0&&chosen<checkboxes.length;all.disabled=publishing||checkboxes.length===0;}
 const count=$('#publish-count');if(count)count.textContent=`Đã chọn ${publishSelection.size} câu`;
 for(const action of ['publish-selected','clear-publish-selection']){const button=$(`[data-action="${action}"]`);if(button){button.dataset.unavailable=String(publishing||publishSelection.size===0);button.disabled=publishing||publishSelection.size===0;}}
}
$('#content').addEventListener('change',event=>{
 const input=event.target;
 if(publishing||(!input.matches('[data-publish-id]')&&input.id!=='select-question-page'))return;
 const targets=input.id==='select-question-page'?$$('[data-publish-id]:not(:disabled)'):[input];
 for(const target of targets){const id=Number(target.dataset.publishId);if(input.checked){if(publishSelection.size>=500&&!publishSelection.has(id)){notify('Mỗi lần xuất bản tối đa 500 câu.');break;}publishSelection.add(id);}else publishSelection.delete(id);}
 updatePublishSelection();
});
async function publishQuestions() {
 if(publishing||publishSelection.size===0)return;
 const ids=[...publishSelection];
 if(!confirm(`Xuất bản ${ids.length} câu đã chọn? Câu hợp lệ sẽ hiển thị cho người học. Câu thiếu lời giải hoặc đáp án hợp lệ sẽ được giữ nguyên và báo lỗi.`))return;
 publishing=true;updatePublishSelection();
 // Keep the selected batch stable while the request is in flight.
 $('#workspace').inert=true;
 try {
  const result=await request('/api/v1/admin/questions/publish','POST',{question_ids:ids});
  if(!session)return;
  publishResult=result;
  for(const id of result.published_ids)publishSelection.delete(id);
  notify(`Đã xuất bản ${result.published_ids.length} câu; ${result.failed.length} câu cần kiểm tra.`);
  await render();
 } finally {publishing=false;$('#workspace').inert=false;updatePublishSelection();}
}
function importPage() {return heading('Nhập dữ liệu','Bổ sung ngân hàng câu hỏi từ file JSON hoặc nguồn crawl được cấu hình.')+`<div class="grid2"><section class="card"><span class="eyebrow">01 / TỪ FILE</span><h2>Import câu hỏi JSON</h2><p class="muted">Tối đa 500 câu và 5 MiB mỗi lần. Một câu lỗi sẽ hủy toàn bộ lô nhập.</p><label>Chọn file JSON<input id="import-file" type="file" accept=".json,application/json"></label><label>Hoặc dán JSON<textarea id="import-json" rows="10" placeholder='{"questions": [...]}'></textarea></label><div class="actions">${actionButton('download-template','Tải mẫu JSON')}${actionButton('import','Kiểm tra và nhập','class="primary"')}</div><p class="muted">Nhập lại file sẽ tạo câu mới; chức năng import không tự chống trùng nội dung.</p></section><section class="card"><span class="eyebrow">02 / TỪ NGUỒN CẤU HÌNH</span><h2>Cào bộ câu hỏi A1 / B</h2><p class="muted">Lấy nội dung từ HocLaixeOnline, tự tạo chủ đề và lưu câu hỏi dưới dạng bản nháp để bạn rà soát.</p><div class="notice">Có chống nhập trùng theo nguồn và hạng. Ảnh được lưu dưới dạng URL. Tác vụ có thể mất thời gian tùy kết nối nguồn.</div>${actionButton('crawl','Bắt đầu cào dữ liệu','class="primary"')}<p class="muted">Sau khi thành công, chờ ít nhất 60 giây trước khi chạy lại.</p></section></div><section class="card"><h3>Kết quả thao tác</h3><pre id="import-result">Chưa chạy tác vụ.</pre></section>`;}
async function apiPage() {
 const data=await request('/admin/api.json');
 return heading('Danh mục REST API',`${data.length} endpoint hiện có · Chú thích chức năng và quyền truy cập.`)+`<div class="notice">Base URL: ${esc(location.origin + backendPrefix)}. Các API cần đăng nhập sử dụng header Authorization: Bearer &lt;access_token&gt;. API legacy được giữ nguyên tên để tương thích.</div>`+ [...new Set(data.map(r=>r.group))].map(group=>`<section class="card"><h3>${esc(group)}</h3><div class="table-wrap"><table><thead><tr><th>Method</th><th>Endpoint</th><th>Quyền</th><th>Chú thích</th></tr></thead><tbody>${data.filter(r=>r.group===group).map(r=>`<tr><td><span class="api-method">${r.method}</span></td><td><code>${esc(r.path)}</code></td><td>${esc(r.access)}</td><td>${esc(r.note)}</td></tr>`).join('')}</tbody></table></div></section>`).join('');
}
$('#content').addEventListener('submit',event=>{if(event.target.id!=='filters')return;event.preventDefault();if(publishing)return;publishSelection.clear();publishResult=null;filters=Object.fromEntries([...new FormData(event.target)].filter(([,v])=>v!==''));page=1;render();});
$('#content').addEventListener('click',event=>{const b=event.target.closest('[data-action]');if(!b)return;busy(b,()=>handleAction(b));});
async function handleAction(b) {
 const action=b.dataset.action,id=Number(b.dataset.id);
 if(action==='publish-selected')return publishQuestions();
 if(action==='clear-publish-selection'){publishSelection.clear();updatePublishSelection();return;}
 if(action==='reload')return render();
 if(action==='prev'||action==='next'){page+=action==='prev'?-1:1;return render();}
 if(action==='go-tests')return navigate('tests');
 if(action==='new-sign')return openSign();
 if(action==='edit-sign')return openSign(await request(`/api/v1/admin/road-signs/${id}`));
 if(action==='new-template')return openTemplate();
 if(action==='edit-template')return openTemplate(await request(`/api/v1/admin/exam-templates/${id}`));
 if(action==='new-question')return openQuestion();
 if(action==='new-category')return openCategory();
 if(action==='new-test')return openTest();
 if(action==='generate')return openTest(null,true);
 if(action==='edit-test')return openTest(await request(`/api/v1/admin/tests/${id}`));
 if(action==='edit'){if(view==='categories')return openCategory(rows.find(r=>r.id===id));return openQuestion(await request(`/api/v1/admin/questions/${id}`));}
 if(action==='delete'){
  if(!confirm(`Xóa ${view==='categories'?'chủ đề':view==='questions'?'câu hỏi':view==='signs'?'biển báo':view==='templates'?'cấu hình':'đề thi'} #${id}? Thao tác này sẽ xóa mềm dữ liệu.`))return;
  await request(`/api/v1/admin/${view==='signs'?'road-signs':view==='templates'?'exam-templates':view}/${id}`,'DELETE');publishSelection.delete(id);notify('Đã xóa dữ liệu.');if(rows.length===1&&page>1)page--;return render();
 }
 if(action==='download-template'){
  const sample={questions:[{title:'Câu hỏi minh họa',content:'Chọn phương án đúng trong ví dụ?',license_type:'A1',category_id:categories[0]?.id||1,image:'',explanation:'Lời giải minh họa — thay bằng nội dung đã kiểm chứng.',is_critical:false,published:false,answers:[{label:'A',content:'Phương án A',order:1,is_correct:true},{label:'B',content:'Phương án B',order:2,is_correct:false}]}]};
  const url=URL.createObjectURL(new Blob([JSON.stringify(sample,null,2)],{type:'application/json'}));const link=document.createElement('a');link.href=url;link.download='cau-hoi-mau.json';link.click();setTimeout(()=>URL.revokeObjectURL(url),1000);return;
 }
 if(action==='import'){
  let raw=$('#import-json').value.trim();const file=$('#import-file').files[0];
  if(file&&raw)throw new Error('Chỉ chọn một nguồn: file JSON hoặc nội dung dán.');
  if(file){if(file.size>5*1024*1024)throw new Error('File vượt quá 5 MiB.');raw=await file.text();}
  if(!raw)throw new Error('Chọn file JSON hoặc dán nội dung trước khi nhập.');
  if(new Blob([raw]).size>5*1024*1024)throw new Error('Nội dung vượt quá 5 MiB.');
  let input;try{input=JSON.parse(raw);}catch{throw new Error('JSON không hợp lệ. Kiểm tra dấu ngoặc và dấu phẩy.');}
  if(!Array.isArray(input?.questions)||input.questions.length<1||input.questions.length>500)throw new Error('JSON phải có mảng questions gồm 1–500 câu.');
  if(!confirm(`Nhập ${input.questions.length} câu hỏi? Nhập lại cùng file sẽ tạo thêm câu mới.`))return;
  const out=await request('/api/v1/admin/questions/import','POST',input);$('#import-result').textContent=`Đã nhập thành công ${out.imported} câu hỏi.`;$('#import-json').value='';$('#import-file').value='';notify('Import hoàn tất.');return;
 }
 if(action==='crawl'){
  if(!confirm('Cào dữ liệu A1/B và lưu các câu mới dưới dạng bản nháp?'))return;
  const result=$('#import-result');result.textContent='Đang lấy dữ liệu, vui lòng chờ…';
  try{const out=await request('/api/v1/admin/questions/crawl','POST');if(result.isConnected)result.textContent=JSON.stringify(out,null,2);categories=(await request('/api/v1/categories')).data||[];notify('Cào dữ liệu hoàn tất.');}catch(e){if(result.isConnected)result.textContent=e.message;throw e;}
 }
}
function openDialog(title,body,kind,id=0) {
 editing={kind,id};$('#editor-title').textContent=title;$('#editor-body').innerHTML=body;$('#editor-error').textContent='';$('#editor').showModal();
}
$$('[data-close]').forEach(b=>b.onclick=()=>$('#editor').close());
function openCategory(c={}) {openDialog(c.id?'Sửa chủ đề':'Thêm chủ đề',field('name','Tên chủ đề',c.name,'text','required maxlength="255"')+area('description','Mô tả',c.description,'maxlength="255"')+`<label>Loại chủ đề<select name="type">${options(categoryTypes,c.type||'UNCATEGORIZED')}</select></label>`,'categories',c.id);}
function answerRow(a={},index=0) {return `<div class="answer-row"><input type="radio" name="correct_answer" aria-label="Đáp án đúng" ${a.is_correct?'checked':''} required><input class="answer-label" aria-label="Nhãn đáp án" maxlength="2" value="${esc(a.label||String.fromCharCode(65+index))}" required><input class="answer-content" aria-label="Nội dung đáp án" placeholder="Nội dung phương án" value="${esc(a.content)}" required><button type="button" data-dialog="remove-answer" aria-label="Xóa phương án">×</button></div>`;}
async function openQuestion(q={}) {
 if(!categories.length)throw new Error('Hãy tạo ít nhất một chủ đề trước khi thêm câu hỏi.');
 openDialog(q.id?`Sửa câu hỏi #${q.id}`:'Thêm câu hỏi',`<div class="form-grid"><div class="span2">${field('title','Tiêu đề',q.title,'text','required')}${area('content','Nội dung câu hỏi',q.content,'required')}</div><label>Chủ đề<select name="category_id">${catOptions(q.category_id)}</select></label><label>Hạng bằng<select name="license_type">${options(licenses,q.license_type||'A1')}</select></label><div class="span2">${field('image','URL hoặc đường dẫn ảnh (không bắt buộc)',q.image,'text')}<img id="image-preview" class="preview" alt="Ảnh minh họa câu hỏi" hidden><label>Liên kết biển báo (không bắt buộc)<select name="road_sign_id"><option value="">Không liên kết</option></select></label>${area('explanation','Lời giải (bắt buộc khi xuất bản)',q.explanation)}<h3>Phương án trả lời</h3><p class="muted">Chọn đúng một đáp án đúng. Thứ tự hiển thị là thứ tự phương án.</p><div id="answers">${(q.answers||[{is_correct:true},{}]).map(answerRow).join('')}</div>${actionButton('unused','＋ Thêm phương án','type="button" data-dialog="add-answer"')}${check('is_critical','Đây là câu điểm liệt',q.is_critical)}${check('published','Xuất bản cho người học',q.published)}${q.source_url?`<p class="muted">Nguồn: ${esc(q.source_url)}</p>`:''}</div></div>`,'questions',q.id);
 const imageInput=$('[name=image]',$('#editor'));
 const preview=()=>{const url=imageInput.value.trim();const img=$('#image-preview');if(/^https?:\/\//i.test(url)||/^\/(?!\/)/.test(url)){img.src=url;img.hidden=false;}else{img.removeAttribute('src');img.hidden=true;}};
 imageInput.onchange=preview;$('#image-preview').onerror=()=>{$('#image-preview').hidden=true;};preview();
 await loadSignOptions(q.road_sign_id);
 for(const name of ['category_id','license_type'])$(`[name=${name}]`,$('#editor')).onchange=()=>loadSignOptions().catch(e=>$('#editor-error').textContent=e.message);
}
function testFields(t={}) {return `<div class="form-grid"><div class="span2">${field('name','Tên đề',t.name,'text','required maxlength="100"')}</div><label>Hạng bằng<select name="license_type">${options(licenses,t.license_type||'A1')}</select></label><label>Loại đề<select name="category_type">${options(categoryTypes,t.category_type,'Tổng hợp (mọi loại)')}</select></label>${field('version','Phiên bản',t.version||'v1','text','required')}${field('duration_minutes','Thời gian (phút)',t.duration_minutes||19,'number','required min="1" max="180"')}${field('passing_score','Điểm đạt',t.passing_score||1,'number','required min="1" max="200"')}<div class="span2">${area('description','Mô tả',t.description,'maxlength="255"')}${check('critical_fail','Trượt nếu sai hoặc bỏ trống câu điểm liệt',t.critical_fail??true)}${check('published','Xuất bản đề cho người học',t.published)}</div></div>`;}
async function openTest(t=null,generate=false) {
 if(!categories.length)throw new Error('Hãy tạo chủ đề và câu hỏi trước khi tạo đề.');
 selected=new Map((t?.questions||[]).map(link=>[link.question_id,link.question?.title||`Câu #${link.question_id}`]));pickerPage=1;
 let extra=generate?`<h3>Số lượng theo chủ đề</h3><p class="muted">Câu điểm liệt nằm trong tổng số câu của mỗi chủ đề.</p><div id="quotas">${(t?.quotas||[{}]).map(quotaRow).join('')}</div><button type="button" data-dialog="add-quota">＋ Thêm chủ đề</button>`:`<h3>Câu hỏi trong đề</h3><p class="muted">Chọn 1–200 câu đã xuất bản, cùng hạng bằng. Dùng mũi tên để sắp xếp thứ tự.</p><ol id="selected-list" class="selected-list"></ol><div class="picker"><div class="toolbar"><label>Tìm câu hỏi<input id="picker-search" placeholder="Nhập từ khóa"></label><button type="button" data-dialog="search-picker">Tìm</button></div><div id="picker-list" class="picker-list"></div><div class="pager"><button type="button" data-dialog="picker-prev">← Trước</button><span id="picker-page"></span><button type="button" data-dialog="picker-next">Tiếp →</button></div></div>`;
 openDialog(generate?'Sinh đề theo chủ đề':t?`Sửa đề #${t.id}`:'Tạo đề thi',testFields(t||{})+extra,generate?'generate':'tests',t?.id);
 if(!generate){renderSelected();await loadPicker();for(const name of ['license_type','category_type'])$(`[name=${name}]`,$('#editor')).onchange=()=>{selected.clear();renderSelected();pickerPage=1;loadPicker().catch(e=>$('#editor-error').textContent=e.message);};}
}
function quotaRow(q={}) {return `<div class="quota-row"><label>Chủ đề<select class="quota-category">${catOptions(q.category_id)}</select></label><label>Tổng câu<input class="quota-count" type="number" value="${esc(q.count??1)}" min="1" max="200" required></label><label>Câu liệt<input class="quota-critical" type="number" value="${esc(q.critical_count??0)}" min="0" max="200" required></label><button type="button" data-dialog="remove-quota" aria-label="Xóa quota">×</button></div>`;}
function renderSelected() {
 $('#selected-list').innerHTML=[...selected].map(([id,title])=>`<li>${esc(title)} <small>#${id}</small><button type="button" data-dialog="move-up" data-id="${id}" aria-label="Đưa lên">↑</button><button type="button" data-dialog="move-down" data-id="${id}" aria-label="Đưa xuống">↓</button><button type="button" data-dialog="unselect" data-id="${id}">Bỏ</button></li>`).join('')||'<li class="muted">Chưa chọn câu hỏi.</li>';
}
let pickerVersion=0;
async function loadPicker() {
 const version=++pickerVersion;
 const params=new URLSearchParams({license_type:$('[name=license_type]',$('#editor')).value,category_type:$('[name=category_type]',$('#editor')).value,search:$('#picker-search').value,page:pickerPage,limit:20});
 $('#picker-list').textContent='Đang tải…';
 // The practice endpoint filters published questions before COUNT/LIMIT/OFFSET.
 // Filtering an admin page in the browser creates empty pages full of drafts.
 const result=await request(`/api/v1/practice/questions?${params}`);
 if(version!==pickerVersion||!$('#picker-list'))return;
 const available=result.data||[];
 $('#picker-list').innerHTML=available.map(q=>`<label><input type="checkbox" data-pick="${q.id}" data-title="${esc(q.title)}" ${selected.has(q.id)?'checked':''}><span>#${q.id} · ${esc(q.title)} ${q.is_critical?'<span class="badge critical">Điểm liệt</span>':''}<br><small>${esc(q.content)}</small></span></label>`).join('')||'<p class="muted">Không có câu đã xuất bản phù hợp. Thử đổi từ khóa hoặc xuất bản câu hỏi trước.</p>';
 $('#picker-page').textContent=`Trang ${pickerPage}/${Math.max(1,Math.ceil(result.total/20))} · ${result.total} câu đã xuất bản`;
 for (const [direction, unavailable] of [['prev',pickerPage<=1],['next',pickerPage*20>=result.total]]) { const button=$(`[data-dialog=picker-${direction}]`); button.dataset.unavailable=String(unavailable); button.disabled=unavailable; }
}
$('#editor-body').addEventListener('change',event=>{if(!event.target.matches('[data-pick]'))return;const b=event.target,id=Number(b.dataset.pick);if(b.checked)selected.set(id,b.dataset.title);else selected.delete(id);renderSelected();});
$('#editor-body').addEventListener('click',event=>{const b=event.target.closest('[data-dialog]');if(!b)return;busy(b,async()=>{
 const action=b.dataset.dialog,id=Number(b.dataset.id);
 if(action==='add-answer')$('#answers').insertAdjacentHTML('beforeend',answerRow({},$$('.answer-row').length));
 if(action==='remove-answer'){if($$('.answer-row').length<=2)throw new Error('Cần ít nhất hai phương án.');b.closest('.answer-row').remove();}
 if(action==='add-quota')$('#quotas').insertAdjacentHTML('beforeend',quotaRow());
 if(action==='remove-quota'){if($$('.quota-row').length<=1)throw new Error('Cần ít nhất một chủ đề.');b.closest('.quota-row').remove();}
 if(action==='search-picker'){pickerPage=1;await loadPicker();}
 if(action==='picker-prev'||action==='picker-next'){pickerPage+=action==='picker-prev'?-1:1;await loadPicker();}
 if(action==='unselect'){selected.delete(id);renderSelected();const checkbox=$(`[data-pick="${id}"]`);if(checkbox)checkbox.checked=false;}
 if(action==='move-up'||action==='move-down'){const entries=[...selected],index=entries.findIndex(([key])=>key===id),target=index+(action==='move-up'?-1:1);if(target>=0&&target<entries.length){[entries[index],entries[target]]=[entries[target],entries[index]];selected=new Map(entries);renderSelected();}}
 },$('#editor-error'));});
$('#editor-form').addEventListener('submit',event=>{event.preventDefault();busy($('button[type=submit]',event.target),async()=>{
 $('#editor-error').textContent='';const input=Object.fromEntries(new FormData(event.target));const {kind,id}=editing;
 let body;
 if(kind==='categories')body={name:input.name.trim(),description:input.description.trim(),type:input.type};
 if(kind==='questions'){
  body={road_sign_id:input.road_sign_id?Number(input.road_sign_id):null,title:input.title.trim(),content:input.content.trim(),image:input.image.trim(),explanation:input.explanation.trim(),category_id:Number(input.category_id),license_type:input.license_type,is_critical:!!input.is_critical,published:!!input.published,answers:$$('.answer-row').map((r,i)=>({label:$('.answer-label',r).value.trim(),content:$('.answer-content',r).value.trim(),order:i+1,is_correct:$('input[type=radio]',r).checked}))};
  if(body.published&&!body.explanation)throw new Error('Thêm lời giải trước khi xuất bản câu hỏi.');
  if(body.answers.filter(a=>a.is_correct).length!==1)throw new Error('Chọn đúng một đáp án đúng.');
 }
 if(kind==='road-signs')body={...input,category_id:Number(input.category_id),published:!!input.published};
 if(kind==='tests'||kind==='generate'||kind==='exam-templates'){
  body={category_type:input.category_type,name:input.name.trim(),description:input.description.trim(),license_type:input.license_type,version:input.version.trim(),duration_minutes:Number(input.duration_minutes),passing_score:Number(input.passing_score),critical_fail:!!input.critical_fail,published:!!input.published};
  if(kind==='tests'){body.question_ids=[...selected.keys()];if(body.question_ids.length<1||body.question_ids.length>200)throw new Error('Chọn từ 1 đến 200 câu hỏi.');if(body.passing_score>body.question_ids.length)throw new Error('Điểm đạt không được lớn hơn số câu.');}
  else body.quotas=$$('.quota-row').map(r=>({category_id:Number($('.quota-category',r).value),count:Number($('.quota-count',r).value),critical_count:Number($('.quota-critical',r).value)}));
 }
 const path=kind==='generate'?'/api/v1/admin/tests/generate':`/api/v1/admin/${kind}${id?'/'+id:''}`;
 await request(path,id?'PUT':'POST',body);$('#editor').close();notify('Đã lưu dữ liệu thành công.');
 if(kind==='categories')categories=(await request('/api/v1/categories')).data||[];
 await render();
 },$('#editor-error'));});
if(session)enter().catch(error=>{logout();$('#login-error').textContent=error.message;});

async function mobileAdminList(kind){
 const endpoint=kind==='signs'?'road-signs':'exam-templates';
 const out=await request(`/api/v1/admin/${endpoint}?${new URLSearchParams({page,limit:20,...filters})}`);rows=out.data||[];
 const sign=kind==='signs';
 return heading(titles[kind],sign?'Học liệu biển báo theo từng hạng bằng; liên kết câu hỏi để luyện và thi.':'Cấu hình do admin duyệt, dùng khi learner bấm Đề ngẫu nhiên.',actionButton(sign?'new-sign':'new-template',sign?'＋ Thêm biển báo':'＋ Thêm cấu hình','class="primary"'))+
 `<form id="filters" class="toolbar"><label>Tìm kiếm<input name="search" value="${esc(filters.search||'')}"></label><label>Hạng bằng<select name="license_type">${options(licenses,filters.license_type,'Tất cả hạng')}</select></label><label>Trạng thái<select name="published">${options([['true','Đã xuất bản'],['false','Bản nháp']],filters.published,'Tất cả')}</select></label>${sign?`<label>Nhóm biển<select name="sign_type">${options(signTypes,filters.sign_type,'Tất cả nhóm')}</select></label>`:`<label>Loại chủ đề<select name="category_type">${options(categoryTypes,filters.category_type,'Tất cả loại')}</select></label>`}<button>Lọc dữ liệu</button></form>`+
 `<div class="table-wrap"><table><thead><tr><th>Tên / Mã</th><th>Hạng / Loại</th><th>Trạng thái</th><th>Thao tác</th></tr></thead><tbody>${rows.map(item=>`<tr><td class="title"><strong>${esc(item.name)}</strong><small>#${item.id} · ${esc(sign?item.code:item.config.version)}</small></td><td>${esc(item.license_type)} · ${esc(sign?item.sign_type:item.category_type||'Tổng hợp')}</td><td>${badge(item.published)}</td><td><div class="row-actions">${actionButton(sign?'edit-sign':'edit-template','Sửa',`data-id="${item.id}"`)}${actionButton('delete','Xóa',`data-id="${item.id}" class="danger"`)}</div></td></tr>`).join('')||'<tr><td colspan="4" class="empty">Chưa có dữ liệu phù hợp.</td></tr>'}</tbody></table></div>`+pager(out.total);
}
function openSign(sign={}){
 const cats=categories.filter(c=>c.type==='SIGN');
 if(!cats.length)throw new Error('Tạo hoặc sửa chủ đề với loại SIGN trước khi thêm biển báo.');
 openDialog(sign.id?'Sửa biển báo':'Thêm biển báo',`<div class="form-grid">${field('code','Mã biển',sign.code,'text','required maxlength="50"')}${field('name','Tên biển',sign.name,'text','required maxlength="255"')}<label>Hạng bằng<select name="license_type">${options(licenses,sign.license_type||'A1')}</select></label><label>Nhóm biển<select name="sign_type">${options(signTypes,sign.sign_type||'WARNING')}</select></label><label>Chủ đề SIGN<select name="category_id">${options(cats.map(c=>[c.id,c.name]),sign.category_id)}</select></label><div class="span2">${field('image','URL / đường dẫn ảnh',sign.image)}${area('meaning','Ý nghĩa',sign.meaning)}${area('description','Ghi chú học tập',sign.description)}${check('published','Xuất bản học liệu (cần ảnh và ý nghĩa)',sign.published)}</div></div>`,'road-signs',sign.id);
}
async function openTemplate(template){
 await openTest(template?.config||null,true);
 editing={kind:'exam-templates',id:template?.id||0};
 $('#editor-title').textContent=template?'Sửa cấu hình đề ngẫu nhiên':'Cấu hình đề ngẫu nhiên cho learner';
}
let signOptionsVersion=0;
async function loadSignOptions(current){
 const version=++signOptionsVersion;
 const form=$('#editor-form');const license=$('[name=license_type]',form).value,category=$('[name=category_id]',form).value;
 let all=[],p=1,total=1;
 while(all.length<total){const result=await request(`/api/v1/admin/road-signs?license_type=${encodeURIComponent(license)}&category_id=${category}&page=${p++}&limit=100`);all.push(...result.data);total=result.total;if(!result.data.length)break;}
 if(version!==signOptionsVersion||!$('[name=road_sign_id]',form))return;
 $('[name=road_sign_id]',form).innerHTML=options(all.map(s=>[s.id,`${s.code} · ${s.name}${s.published?'':' (nháp)'}`]),current,'Không liên kết');
}
