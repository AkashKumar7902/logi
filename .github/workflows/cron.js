const https = require('https');

const backendUrl = process.env.LOGI_KEEP_ALIVE_URL || 'https://logi-y295.onrender.com/healthz';
const timeoutMs = Number(process.env.LOGI_KEEP_ALIVE_TIMEOUT_MS || 10000);

console.log(`Pinging backend health endpoint: ${backendUrl}`);

const request = https
  .get(backendUrl, (res) => {
    res.resume();
    if (res.statusCode >= 200 && res.statusCode < 300) {
      console.log(`Backend is healthy: ${res.statusCode}`);
    } else {
      console.error(`Backend health check failed with status code: ${res.statusCode}`);
      process.exitCode = 1;
    }
  })
  .on('error', (err) => {
    console.error('Backend health check error:', err.message);
    process.exitCode = 1;
  });

request.setTimeout(timeoutMs, () => {
  request.destroy(new Error(`Backend health check timed out after ${timeoutMs}ms`));
});
