import { initializeApp } from 'firebase/app';
import { getAuth } from 'firebase/auth';
import { getFirestore } from 'firebase/firestore';
import { getStorage } from 'firebase/storage';

// Your web app's Firebase configuration
const firebaseConfig = {
  apiKey: "AIzaSyBUIFUtzrY5YvhLmGQsl4Vz2IpnA_rNYd8",
  authDomain: "elearning-raisa.firebaseapp.com",
  projectId: "elearning-raisa",
  storageBucket: "elearning-raisa.appspot.com",
  messagingSenderId: "196790250714",
  appId: "1:196790250714:web:18a05fd86e013c708a89e3"
};


// Initialize Firebase
const app = initializeApp(firebaseConfig);
export const auth = getAuth(app);
export const db = getFirestore(app);
export const storage = getStorage(app);

export default app;
