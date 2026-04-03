package server

import "net/http"

func (s *Server) dashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(dashHTML))
}

const dashHTML = `<!DOCTYPE html>
<html lang="en"><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>Archive</title>
<style>
:root{--bg:#1a1410;--bg2:#241e18;--bg3:#2e261e;--rust:#c45d2c;--rl:#e8753a;--leather:#a0845c;--ll:#c4a87a;--cream:#f0e6d3;--cd:#bfb5a3;--cm:#7a7060;--gold:#d4a843;--green:#4a9e5c;--red:#c44040;--blue:#4a7ec4;--mono:'JetBrains Mono',Consolas,monospace;--serif:'Libre Baskerville',Georgia,serif}
*{margin:0;padding:0;box-sizing:border-box}body{background:var(--bg);color:var(--cream);font-family:var(--mono);font-size:13px;line-height:1.6;height:100vh;overflow:hidden}
a{color:var(--rl);text-decoration:none}a:hover{color:var(--gold)}
.app{display:flex;height:100vh}

.sidebar{width:200px;background:var(--bg2);border-right:1px solid var(--bg3);display:flex;flex-direction:column;flex-shrink:0;overflow-y:auto}
.sidebar-hdr{padding:.6rem .8rem;border-bottom:1px solid var(--bg3);display:flex;justify-content:space-between;align-items:center}
.sidebar-hdr span{font-family:var(--serif);font-size:.9rem;color:var(--rl)}
.sidebar-section{padding:.3rem .8rem;font-size:.6rem;text-transform:uppercase;letter-spacing:1.5px;color:var(--rust);margin-top:.5rem}
.sidebar-item{padding:.25rem .8rem;font-size:.73rem;cursor:pointer;display:flex;align-items:center;gap:.4rem;transition:.1s;color:var(--cd)}
.sidebar-item:hover{background:var(--bg3)}.sidebar-item.active{background:var(--bg3);color:var(--cream)}
.sidebar-count{margin-left:auto;font-size:.6rem;color:var(--cm)}
.sidebar-bottom{margin-top:auto;padding:.5rem .8rem;border-top:1px solid var(--bg3);font-size:.6rem;color:var(--cm)}

.list-pane{width:300px;border-right:1px solid var(--bg3);display:flex;flex-direction:column;flex-shrink:0}
.list-toolbar{padding:.4rem .6rem;border-bottom:1px solid var(--bg3);display:flex;gap:.4rem;align-items:center}
.list-toolbar input{flex:1;background:var(--bg);border:1px solid var(--bg3);color:var(--cream);padding:.3rem .5rem;font-family:var(--mono);font-size:.72rem;outline:none}
.list-toolbar input:focus{border-color:var(--rust)}
.list-scroll{flex:1;overflow-y:auto}
.clip-item{padding:.5rem .7rem;border-bottom:1px solid var(--bg3);cursor:pointer;transition:.1s}
.clip-item:hover{background:var(--bg2)}.clip-item.active{background:var(--bg2);border-left:2px solid var(--rl)}
.clip-title{font-size:.78rem;font-weight:600;display:flex;align-items:center;gap:.3rem}
.clip-title .fav{color:var(--gold);font-size:.65rem}
.clip-excerpt{font-size:.68rem;color:var(--cm);margin-top:.15rem;overflow:hidden;text-overflow:ellipsis;display:-webkit-box;-webkit-line-clamp:2;-webkit-box-orient:vertical}
.clip-meta{font-size:.6rem;color:var(--cm);margin-top:.15rem;display:flex;gap:.5rem}
.clip-tag{font-size:.55rem;padding:0 .25rem;background:var(--bg3);color:var(--ll);border-radius:2px}
.status-dot{width:6px;height:6px;border-radius:50%;flex-shrink:0}.st-inbox{background:var(--blue)}.st-read{background:var(--green)}.st-archived{background:var(--cm)}

.reader{flex:1;display:flex;flex-direction:column;min-width:0}
.reader-toolbar{padding:.4rem .8rem;border-bottom:1px solid var(--bg3);display:flex;align-items:center;gap:.5rem}
.btn{font-family:var(--mono);font-size:.68rem;padding:.25rem .6rem;border:1px solid;cursor:pointer;background:transparent;transition:.15s;white-space:nowrap}
.btn-p{border-color:var(--rust);color:var(--rl)}.btn-p:hover{background:var(--rust);color:var(--cream)}
.btn-d{border-color:var(--bg3);color:var(--cm)}.btn-d:hover{border-color:var(--red);color:var(--red)}
.btn-s{border-color:var(--green);color:var(--green)}.btn-s:hover{background:var(--green);color:var(--bg)}
.btn-g{border-color:var(--gold);color:var(--gold)}.btn-g:hover{background:var(--gold);color:var(--bg)}
.reader-content{flex:1;overflow-y:auto;padding:1.2rem 2rem;max-width:700px}
.reader-content h1{font-family:var(--serif);font-size:1.3rem;margin-bottom:.3rem}
.reader-content .reader-url{font-size:.7rem;color:var(--leather);margin-bottom:.5rem;word-break:break-all}
.reader-content .reader-meta{font-size:.68rem;color:var(--cm);margin-bottom:1rem;display:flex;gap:.8rem}
.reader-content .reader-body{font-family:var(--serif);font-size:.88rem;line-height:1.9;color:var(--cd)}
.reader-content .reader-body p{margin:.6rem 0}
.annotations{border-top:1px solid var(--bg3);max-height:180px;overflow-y:auto;padding:.5rem .8rem}
.ann-item{padding:.3rem 0;border-bottom:1px solid var(--bg3);font-size:.72rem}
.ann-highlight{border-left:3px solid;padding-left:.4rem;color:var(--cream);font-style:italic;margin-bottom:.2rem}
.ann-note{color:var(--cd)}

.empty{display:flex;align-items:center;justify-content:center;flex:1;color:var(--cm);font-style:italic;font-family:var(--serif)}

.modal-bg{position:fixed;top:0;left:0;right:0;bottom:0;background:rgba(0,0,0,.65);display:flex;align-items:center;justify-content:center;z-index:100}
.modal{background:var(--bg2);border:1px solid var(--bg3);padding:1.5rem;width:95%;max-width:550px;max-height:90vh;overflow-y:auto}
.modal h2{font-family:var(--serif);font-size:.95rem;margin-bottom:1rem}
label.fl{display:block;font-size:.65rem;color:var(--leather);text-transform:uppercase;letter-spacing:1px;margin-bottom:.25rem;margin-top:.7rem}
input[type=text],textarea,select{background:var(--bg);border:1px solid var(--bg3);color:var(--cream);padding:.4rem .6rem;font-family:var(--mono);font-size:.8rem;width:100%;outline:none}
input:focus,textarea:focus{border-color:var(--rust)}
textarea{resize:vertical;min-height:80px}
</style>
<link href="https://fonts.googleapis.com/css2?family=Libre+Baskerville:ital@0;1&family=JetBrains+Mono:wght@400;600&display=swap" rel="stylesheet">
</head><body>
<div class="app">
<div class="sidebar">
<div class="sidebar-hdr"><span>Archive</span><button class="btn btn-p" style="font-size:.6rem;padding:.15rem .4rem" onclick="showSaveClip()">+ Save</button></div>
<div class="sidebar-section">Status</div>
<div class="sidebar-item active" data-filter="all" onclick="setFilter('all')">All <span class="sidebar-count" id="sAll">-</span></div>
<div class="sidebar-item" data-filter="inbox" onclick="setFilter('inbox')">📥 Inbox <span class="sidebar-count" id="sInbox">-</span></div>
<div class="sidebar-item" data-filter="read" onclick="setFilter('read')">✓ Read <span class="sidebar-count" id="sRead">-</span></div>
<div class="sidebar-item" data-filter="archived" onclick="setFilter('archived')">📦 Archived <span class="sidebar-count" id="sArchived">-</span></div>
<div class="sidebar-item" data-filter="favorites" onclick="setFilter('favorites')">⭐ Favorites <span class="sidebar-count" id="sFav">-</span></div>
<div class="sidebar-section">Collections</div>
<div id="collList"></div>
<div class="sidebar-item" style="color:var(--rl)" onclick="showNewCollection()">+ Collection</div>
<div class="sidebar-section">Tags</div>
<div id="tagList"></div>
<div class="sidebar-bottom" id="sReadTime">-</div>
</div>

<div class="list-pane">
<div class="list-toolbar">
<input type="text" id="searchBox" placeholder="Search clips..." onkeydown="if(event.key==='Enter')loadClips()">
</div>
<div class="list-scroll" id="clipList"></div>
</div>

<div class="reader" id="readerArea">
<div class="empty" id="emptyReader">Select a clip to read.</div>
</div>
</div>
<div id="modal"></div>

<script>
let clips=[],collections=[],tags=[],curClip=null,curFilter='all',curCollection='',curTag='';

async function api(url,opts){const r=await fetch(url,opts);return r.json()}
function esc(s){return String(s||'').replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;')}
function timeAgo(d){if(!d)return'';const s=Math.floor((Date.now()-new Date(d))/1e3);if(s<60)return s+'s ago';if(s<3600)return Math.floor(s/60)+'m ago';if(s<86400)return Math.floor(s/3600)+'h ago';return Math.floor(s/86400)+'d ago'}

async function init(){
  const[cd,td,sd]=await Promise.all([api('/api/collections'),api('/api/tags'),api('/api/stats')]);
  collections=cd.collections||[];tags=td.tags||[];
  document.getElementById('sAll').textContent=sd.total;
  document.getElementById('sInbox').textContent=sd.inbox;
  document.getElementById('sRead').textContent=sd.read;
  document.getElementById('sArchived').textContent=sd.archived;
  document.getElementById('sFav').textContent=sd.favorites;
  document.getElementById('sReadTime').textContent=sd.total_read_min+' min of reading saved';
  document.getElementById('collList').innerHTML=collections.map(c=>'<div class="sidebar-item" onclick="filterCollection(\''+c.id+'\')">'+esc(c.icon)+' '+esc(c.name)+'<span class="sidebar-count">'+c.clip_count+'</span></div>').join('');
  document.getElementById('tagList').innerHTML=(tags||[]).slice(0,10).map(t=>'<div class="sidebar-item" onclick="filterTag(\''+esc(t)+'\')">#'+esc(t)+'</div>').join('');
  loadClips();
}

function setFilter(f){curFilter=f;curCollection='';curTag='';document.querySelectorAll('.sidebar-item').forEach(el=>el.classList.toggle('active',el.dataset.filter===f));loadClips()}
function filterCollection(id){curFilter='';curCollection=id;curTag='';loadClips()}
function filterTag(t){curFilter='';curCollection='';curTag=t;loadClips()}

async function loadClips(){
  const p=new URLSearchParams();
  if(curFilter==='favorites')p.set('favorite','true');
  else if(curFilter&&curFilter!=='all')p.set('status',curFilter);
  if(curCollection)p.set('collection_id',curCollection);
  if(curTag)p.set('tag',curTag);
  const q=document.getElementById('searchBox').value;if(q)p.set('search',q);
  const d=await api('/api/clips?'+p);clips=d.clips||[];
  renderClips();
}

function renderClips(){
  const el=document.getElementById('clipList');
  if(!clips.length){el.innerHTML='<div style="padding:2rem;text-align:center;color:var(--cm);font-style:italic;font-family:var(--serif)">No clips found.</div>';return}
  el.innerHTML=clips.map(c=>{
    const tags=(c.tags||[]).map(t=>'<span class="clip-tag">'+esc(t)+'</span>').join(' ');
    const active=curClip&&curClip.id===c.id?'active':'';
    return '<div class="clip-item '+active+'" onclick="openClip(\''+c.id+'\')">'+
      '<div class="clip-title"><span class="status-dot st-'+c.status+'"></span>'+(c.favorite?'<span class="fav">⭐</span>':'')+esc(c.title)+'</div>'+
      '<div class="clip-excerpt">'+esc(c.excerpt)+'</div>'+
      '<div class="clip-meta">'+(c.site_name?'<span>'+esc(c.site_name)+'</span>':'')+
        '<span>'+c.read_time+'m</span>'+
        '<span>'+timeAgo(c.created_at)+'</span>'+tags+'</div></div>'
  }).join('')
}

async function openClip(id){
  const[c,ad]=await Promise.all([api('/api/clips/'+id),api('/api/clips/'+id+'/annotations')]);
  curClip=c;
  const anns=(ad.annotations||[]).map(a=>
    '<div class="ann-item">'+(a.highlight?'<div class="ann-highlight" style="border-color:'+esc(a.color)+'">'+esc(a.highlight)+'</div>':'')+
    '<div class="ann-note">'+esc(a.note)+'</div>'+
    '<span style="font-size:.6rem;color:var(--cm)">'+timeAgo(a.created_at)+' <span style="cursor:pointer;color:var(--red)" onclick="delAnn(\''+a.id+'\')">del</span></span></div>').join('');

  const statusBtns={inbox:'<button class="btn btn-s" onclick="setClipStatus(\'read\')">Mark read</button> <button class="btn btn-d" onclick="setClipStatus(\'archived\')">Archive</button>',
    read:'<button class="btn btn-d" onclick="setClipStatus(\'inbox\')">Back to inbox</button> <button class="btn btn-d" onclick="setClipStatus(\'archived\')">Archive</button>',
    archived:'<button class="btn btn-s" onclick="setClipStatus(\'inbox\')">Unarchive</button>'};

  document.getElementById('readerArea').innerHTML=
    '<div class="reader-toolbar">'+
      (statusBtns[c.status]||'')+
      '<button class="btn btn-g" onclick="toggleFav()">'+( c.favorite?'Unfav':'⭐ Fav')+'</button>'+
      '<span style="flex:1"></span>'+
      '<button class="btn btn-d" onclick="if(confirm(\'Delete?\'))delClip(\''+c.id+'\')">Del</button>'+
    '</div>'+
    '<div class="reader-content">'+
      '<h1>'+esc(c.title)+'</h1>'+
      (c.url?'<div class="reader-url"><a href="'+esc(c.url)+'" target="_blank">'+esc(c.url)+'</a></div>':'')+
      '<div class="reader-meta">'+
        (c.author?'<span>'+esc(c.author)+'</span>':'')+
        (c.site_name?'<span>'+esc(c.site_name)+'</span>':'')+
        '<span>'+c.read_time+' min read</span>'+
        '<span>Saved '+timeAgo(c.created_at)+'</span>'+
      '</div>'+
      '<div class="reader-body">'+esc(c.content).split('\n\n').map(p=>'<p>'+p+'</p>').join('')+'</div>'+
    '</div>'+
    '<div class="annotations">'+
      '<div style="font-size:.65rem;color:var(--leather);margin-bottom:.3rem">Annotations ('+c.note_count+')</div>'+
      anns+
      '<div style="display:flex;gap:.3rem;margin-top:.4rem">'+
        '<input type="text" id="annNote" placeholder="Add a note..." style="flex:1;font-size:.68rem;padding:.25rem .4rem">'+
        '<button class="btn btn-p" style="font-size:.6rem" onclick="addAnn()">Add</button>'+
      '</div>'+
    '</div>';
  renderClips();
}

async function setClipStatus(status){if(!curClip)return;await api('/api/clips/'+curClip.id+'/status',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({status})});openClip(curClip.id);init()}
async function toggleFav(){if(!curClip)return;await api('/api/clips/'+curClip.id+'/favorite',{method:'POST'});openClip(curClip.id);init()}
async function delClip(id){await api('/api/clips/'+id,{method:'DELETE'});curClip=null;document.getElementById('readerArea').innerHTML='<div class="empty">Clip deleted.</div>';loadClips();init()}
async function addAnn(){if(!curClip)return;const note=document.getElementById('annNote').value;if(!note)return;await api('/api/clips/'+curClip.id+'/annotations',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({note})});openClip(curClip.id)}
async function delAnn(id){await api('/api/annotations/'+id,{method:'DELETE'});openClip(curClip.id)}

function showSaveClip(){
  const collOpts=collections.map(c=>'<option value="'+c.id+'">'+esc(c.name)+'</option>').join('');
  document.getElementById('modal').innerHTML='<div class="modal-bg" onclick="if(event.target===this)closeModal()"><div class="modal">'+
    '<h2>Save to Archive</h2>'+
    '<label class="fl">URL (optional)</label><input type="text" id="sc-url" placeholder="https://example.com/article">'+
    '<label class="fl">Title</label><input type="text" id="sc-title">'+
    '<label class="fl">Content</label><textarea id="sc-content" rows="5" placeholder="Paste article text..."></textarea>'+
    '<label class="fl">Author</label><input type="text" id="sc-author">'+
    '<label class="fl">Collection</label><select id="sc-coll"><option value="">None</option>'+collOpts+'</select>'+
    '<label class="fl">Tags (comma-separated)</label><input type="text" id="sc-tags" placeholder="engineering, tools">'+
    '<div style="display:flex;gap:.5rem;margin-top:1rem"><button class="btn btn-p" onclick="saveClip()">Save</button><button class="btn btn-d" onclick="closeModal()">Cancel</button></div>'+
  '</div></div>'
}

async function saveClip(){
  const tags=(document.getElementById('sc-tags').value||'').split(',').map(s=>s.trim()).filter(Boolean);
  const body={url:document.getElementById('sc-url').value,title:document.getElementById('sc-title').value,content:document.getElementById('sc-content').value,author:document.getElementById('sc-author').value,collection_id:document.getElementById('sc-coll').value,tags};
  if(!body.title&&!body.url){alert('Title or URL required');return}
  await api('/api/clips',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(body)});
  closeModal();loadClips();init()
}

function showNewCollection(){
  document.getElementById('modal').innerHTML='<div class="modal-bg" onclick="if(event.target===this)closeModal()"><div class="modal">'+
    '<h2>New Collection</h2>'+
    '<label class="fl">Name</label><input type="text" id="nc-name">'+
    '<label class="fl">Icon (emoji)</label><input type="text" id="nc-icon" value="📁" style="width:60px">'+
    '<div style="display:flex;gap:.5rem;margin-top:1rem"><button class="btn btn-p" onclick="saveCollection()">Create</button><button class="btn btn-d" onclick="closeModal()">Cancel</button></div>'+
  '</div></div>'
}

async function saveCollection(){
  const body={name:document.getElementById('nc-name').value,icon:document.getElementById('nc-icon').value};
  if(!body.name){alert('Name required');return}
  await api('/api/collections',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(body)});
  closeModal();init()
}

function closeModal(){document.getElementById('modal').innerHTML=''}
init();
fetch('/api/tier').then(r=>r.json()).then(j=>{if(j.tier==='free'){var b=document.getElementById('upgrade-banner');if(b)b.style.display='block'}}).catch(()=>{var b=document.getElementById('upgrade-banner');if(b)b.style.display='block'});
</script></body></html>`
