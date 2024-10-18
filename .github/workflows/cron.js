// cron.js
const cron = require('cron');
const https = require('https');

const backendUrl = 'https://logi-y295.onrender.com';

const job = new cron.CronJob('*/14 * * * *', function () {
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
});

// Start the cron job
job.start();

module.exports = { job };
