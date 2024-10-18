// cron.js
const https = require('https');

const backendUrl = 'https://logi-y295.onrender.com';

console.log('Attempting to keep server alive...');

https
  .get(backendUrl, (res) => {
    if (res.statusCode === 200) {
      console.log('Server is alive');
    } else {
      console.error(`Failed with status code: ${res.statusCode}`);
    }
  })
  .on('error', (err) => {
    console.error('Error occurred:', err.message);
  });
