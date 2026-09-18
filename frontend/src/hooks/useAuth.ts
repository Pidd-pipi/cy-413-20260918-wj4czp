import {authStore} from '../stores/authStore';export const useAuth=()=>({user:authStore.user,isAuthenticated:Boolean(authStore.token),logout:authStore.logout});
