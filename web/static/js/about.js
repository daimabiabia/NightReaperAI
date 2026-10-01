// NightReaper AI :: About Author modal
function showAboutAuthor() {
    var el = document.getElementById('about-author-modal');
    if (el && typeof openAppModal === 'function') openAppModal(el);
}
function closeAboutAuthor() {
    var el = document.getElementById('about-author-modal');
    if (el && typeof closeAppModal === 'function') closeAppModal(el);
}
window.showAboutAuthor = showAboutAuthor;
window.closeAboutAuthor = closeAboutAuthor;
