const express = require('express');
const bodyParser = require('body-parser');

const app = express();
const port = 3000;

// Parse JSON bodies
app.use(bodyParser.json());

// Parse URL-encoded bodies
app.use(bodyParser.urlencoded({ extended: true }));

// Log all incoming requests
app.use((req, res, next) => {
  console.log('\n=== New Request ===');
  console.log('Method:', req.method);
  console.log('Path:', req.path);
  console.log('Headers:', JSON.stringify(req.headers, null, 2));
  console.log('Body:', JSON.stringify(req.body, null, 2));
  console.log('==================\n');
  next();
});

// Handle all POST requests
app.post('*', (req, res) => {
  res.status(200).json({ message: 'Webhook received successfully' });
});

// Handle all GET requests
app.get('*', (req, res) => {
  res.status(200).json({ message: 'Webhook tester is running' });
});

app.listen(port, () => {
  console.log(`Webhook listener running at http://localhost:${port}`);
  console.log('Ready to receive webhooks!');
}); 