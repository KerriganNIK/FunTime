try {
  var saved = localStorage.getItem('funtime.theme');
  document.documentElement.dataset.theme = saved === 'dark' || saved === 'light' ? saved : 'light';
} catch {
  document.documentElement.dataset.theme = 'light';
}
