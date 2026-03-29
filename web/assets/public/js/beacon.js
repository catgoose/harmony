// setup:feature:demo
/**
 * Lightweight client analytics via navigator.sendBeacon.
 * Logs page views and HTMX navigation events without blocking the user.
 * Fire-and-forget — guaranteed to complete even during page unload.
 */
(function() {
  var endpoint = '/log/beacon';

  function send(event, data) {
    var payload = JSON.stringify({
      event: event,
      path: window.location.pathname,
      referrer: document.referrer || '',
      timestamp: new Date().toISOString(),
      data: data || {}
    });
    navigator.sendBeacon(endpoint, new Blob([payload], { type: 'application/json' }));
  }

  // Log initial page load
  send('page_view');

  // Log HTMX navigations (hx-boost page transitions)
  document.body.addEventListener('htmx:pushedIntoHistory', function(e) {
    send('navigation', { to: e.detail.path });
  });

  // Log page unload with time spent
  var loadTime = Date.now();
  window.addEventListener('pagehide', function() {
    send('page_leave', { duration_ms: Date.now() - loadTime });
  });
})();
