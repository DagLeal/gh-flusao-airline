(() => {
 'use strict';
 let toastTimer, lastFocus;
 const toast = (message) => { const el=document.getElementById('toast'); if(!el)return; el.textContent=message;el.classList.add('visible');clearTimeout(toastTimer);toastTimer=setTimeout(()=>el.classList.remove('visible'),6000); };
 const dirty=()=>document.querySelector('[data-row-id][data-dirty]');
 const updateCount=(grid)=>{const n=grid.querySelectorAll('[data-dirty]').length;const c=grid.querySelector('[data-dirty-count]');if(c)c.textContent=n?n+' linha(s) alterada(s)':'';};
 const closeModal=()=>{const m=document.getElementById('modal');if(m)m.replaceChildren();if(lastFocus)lastFocus.focus();};
 const refresh=()=>{document.querySelectorAll('[data-grid]').forEach(g=>{if(!g.closest('#detail'))htmx.trigger(g,'refreshGrid');});const d=document.querySelector('[data-detail-url]');if(d)htmx.ajax('GET',d.dataset.detailUrl,{target:'#detail',swap:'outerHTML'});};
 document.addEventListener('htmx:configRequest',e=>{const token=document.querySelector('meta[name="csrf-token"]');if(token)e.detail.headers['X-CSRF-Token']=token.content;});
 document.addEventListener('htmx:beforeRequest',e=>{if(e.detail.requestConfig.verb==='get'&&dirty()&&e.detail.target?.id!=='modal'){if(!confirm('Existem alterações não salvas. Deseja descartá-las?'))e.preventDefault();}});
 document.addEventListener('htmx:responseError',e=>{toast(e.detail.xhr.responseText||'Não foi possível concluir a operação. Tente novamente.');});
 document.addEventListener('htmx:sendError',()=>toast('Não foi possível conectar ao servidor. Seus campos foram mantidos.'));
 document.addEventListener('input',e=>{if(e.target.matches('[data-report-search]')){const q=e.target.value.toLocaleLowerCase('pt-BR');document.querySelectorAll('[data-report-item]').forEach(item=>{item.hidden=!item.textContent.toLocaleLowerCase('pt-BR').includes(q);});}});
 document.addEventListener('input',e=>{const cell=e.target.closest('.cell-input');if(!cell)return;const row=cell.closest('[data-row-id]');row.dataset.dirty='true';row.querySelector('[data-changed]').disabled=false;updateCount(row.closest('[data-grid]'));});
 document.addEventListener('change',e=>{if(e.target.matches('[data-inline-toggle]')){const f=e.target.closest('form'),active=e.target.checked;f.querySelector('[data-inline-fields]').hidden=!active;const select=f.querySelector('[name=passageiro_id]');select.required=!active;select.disabled=active;f.querySelectorAll('[name=new_nome],[name=new_documento]').forEach(el=>el.required=active);}});
 document.addEventListener('click',e=>{const button=e.target.closest('[data-seat]');if(button){const f=button.closest('form');f.querySelector('[name=assento]').value=button.dataset.seat;f.querySelectorAll('[data-seat]').forEach(el=>el.classList.remove('chosen'));button.classList.add('chosen');}});
 document.addEventListener('change',e=>{if(e.target.matches('.cell-input'))e.target.dispatchEvent(new Event('input',{bubbles:true}));});
 document.addEventListener('click',e=>{
 const edit=e.target.closest('[data-edit-grid]');if(edit){const g=edit.closest('[data-grid]');g.classList.toggle('editing');const active=g.classList.contains('editing');g.querySelectorAll('[data-save-batch],[data-discard]').forEach(b=>b.hidden=!active);edit.textContent=active?'✓ Modo de edição':'✎ Editar grade';if(active)g.querySelector('.cell-input')?.focus();}
 const discard=e.target.closest('[data-discard]');if(discard){const g=discard.closest('[data-grid]');if(g.querySelector('[data-dirty]')&&!confirm('Descartar as alterações desta grade?'))return;g.querySelector('form.batch-form').reset();g.querySelectorAll('[data-dirty]').forEach(r=>{delete r.dataset.dirty;r.querySelector('[data-changed]').disabled=true;});updateCount(g);g.classList.remove('editing');g.querySelectorAll('[data-save-batch],[data-discard]').forEach(b=>b.hidden=true);g.querySelector('[data-edit-grid]').textContent='✎ Editar grade';}
 if(e.target.closest('[data-close-modal]'))closeModal();
 if(e.target.closest('[data-print]'))window.print();
 const clear=e.target.closest('[data-clear-filters]');if(clear){const form=clear.closest('form');form.querySelectorAll('input:not([type=hidden]),select').forEach(el=>el.value='');form.requestSubmit();}
 if(e.target.closest('[hx-target="#modal"]'))lastFocus=document.activeElement;
 });
 document.addEventListener('keydown',e=>{
 const modal=document.querySelector('[role=dialog]');
 if(e.key==='Escape'&&modal){e.preventDefault();closeModal();}
 if(e.key==='Tab'&&modal){const items=[...modal.querySelectorAll('button,input:not([type=hidden]),select,a[href]')].filter(x=>!x.disabled);const first=items[0],last=items.at(-1);if(e.shiftKey&&document.activeElement===first){e.preventDefault();last.focus();}else if(!e.shiftKey&&document.activeElement===last){e.preventDefault();first.focus();}}
 if(e.target.closest('.batch-form')&&e.key==='Enter'&&e.target.matches('input')){e.preventDefault();e.target.blur();}
 });
 document.addEventListener('htmx:afterSettle',e=>{if(e.detail.target?.id==='modal'){const form=document.querySelector('#modal form');if(form)form.dataset.ready='true';document.querySelector('[role=dialog] input:not([type=hidden]),[role=dialog] select')?.focus();document.querySelector('[data-inline-toggle]:checked')?.dispatchEvent(new Event('change',{bubbles:true}));if(document.querySelector('[data-close-modal]')&&!document.querySelector('[role=dialog]'))closeModal();}});
 document.addEventListener('saved',()=>{closeModal();toast('Registro salvo com sucesso.');refresh();});
 document.addEventListener('unchanged',()=>toast('Nenhuma alteração para salvar.'));
 document.addEventListener('gridSaved',e=>{const form=e.target.closest?.('.batch-form')||e.target;const grid=form.closest?.('[data-grid]');if(grid){grid.querySelectorAll('[data-dirty]').forEach(r=>delete r.dataset.dirty);htmx.trigger(grid,'refreshGrid');}toast('Alterações da grade salvas com sucesso.');});
 document.addEventListener('stateChanged',()=>{toast('Estado atualizado. Os registros relacionados também foram processados.');document.querySelectorAll('[data-grid]').forEach(g=>{if(!g.closest('#detail'))htmx.trigger(g,'refreshGrid');});});
 window.addEventListener('beforeunload',e=>{if(dirty()){e.preventDefault();e.returnValue='';}});
 const entity=new URLSearchParams(location.search).get('entity')||'voo';document.querySelectorAll('.sidebar nav a').forEach(a=>{if((location.pathname==='/consultas-avancadas'&&new URL(a.href).pathname==='/consultas-avancadas')||(location.pathname!=='/consultas-avancadas'&&new URL(a.href).searchParams.get('entity')===entity))a.classList.add('active');});
})();