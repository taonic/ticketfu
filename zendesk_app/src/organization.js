import { createApp } from 'vue';
import OrganizationApp from './components/OrganizationApp.vue';
import './main.css';

// Create and mount the root instance
const app = createApp(OrganizationApp);
app.mount('#app');
