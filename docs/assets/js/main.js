// Add any JavaScript functionality here.

// Lightweight 2000s-flavored interactions
document.addEventListener('DOMContentLoaded', () => {
  // Retro visitor counter (local-only)
  const counterEl = document.getElementById('counter');
  if (counterEl) {
    const key = 'budgie-visitor-count';
    const next = (parseInt(localStorage.getItem(key) || '1', 10) + 1)
      .toString()
      .padStart(5, '0');
    localStorage.setItem(key, next);
    counterEl.textContent = next;
  }

  // Gentle float wobble based on mouse position
  const floaties = document.querySelectorAll('.floating-budgies span');
  document.addEventListener('mousemove', (e) => {
    const { innerWidth, innerHeight } = window;
    const x = (e.clientX / innerWidth - 0.5) * 8;
    const y = (e.clientY / innerHeight - 0.5) * 8;
    floaties.forEach((el, idx) => {
      el.style.transform = `translate(${x * (idx + 1) * 0.2}px, ${y * (idx + 1) * 0.2}px)`;
    });
  });
});