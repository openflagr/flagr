<template>
    <div  class="callback">
        <div>
        <p>{{  text }}</p>
        </div>
        <button v-if="loginBtn" @click="router.push('/login')">
            Go to login
        </button>
    </div>
</template>
  
<script>
    import { ref } from 'vue';
    import { useRouter } from 'vue-router';
    import { googleAuthenticate } from '../utils/apiUtil';
    import { getUserDetails } from '../utils/apiUtil';
    import { useStore } from 'vuex';
    import { logout } from '../utils/apiUtil';
  
  export default {
    name: 'Callback',
    setup() {
      const router = useRouter();
      const userVerified = ref(false);
      const text = ref('Authenticating, please wait...');
      const loginBtn = ref(false)
      const store = useStore();
    //   const userDetails = computed(() => store.state.userDetails);
      const updateValue = (val) => {
            store.dispatch('updateUserDetails', val);
      };
      // This function will handle the callback logic
      const handleAuthCallback = async () => {
        const urlParams = new URLSearchParams(window.location.search);
        const code = urlParams.get('code');
        const state = urlParams.get('state');
  
        if (!code || !state) {
          console.error('Missing code or state in callback URL');
          router.push('/login');  // Redirect to login if parameters are missing
          return;
        }
        const authOrigin = window.location.origin

        const data = {
            state: state,
            google_auth_code: code,
            uiRedirectUrl: `${authOrigin}/callback`,
        }
        if (state) {
            // base64 decode the state
            const parsedState = atob(state)

            // parse query string from state into an object
            const parsedStateQueryParams = Object.fromEntries(new URLSearchParams(parsedState))
            const redirectOrigin = parsedStateQueryParams.redirectOrigin
            // redirect to redirectOrigin if it is a vercel origin with code and state
            
        }
        googleAuthenticate(data).then((res) => {
            if (res.error) {
                userVerified.value = false
                text.value = 'Something went wrong, please try again later, go to login'
                loginBtn.value = true
            } else {
                userVerified.value = true
                getUserDetails().then((res) => {
                    updateValue(res)
                    store.dispatch('updateAuth', true);
                }).catch((err) => {
                    store.dispatch('updateAuth', false);
                    logout()
                })
                router.push('/')
            }
        }).catch(() => {
            console.log("errorororor")
            userVerified.value = false
            text.value = 'Something went wrong, please try again later, go to login'
            loginBtn.value = true
            store.dispatch('updateAuth', false);
        })
      };
  
      // Call the function to handle the callback when the component mounts
      handleAuthCallback();
      return {
        text,
        loginBtn,
        router, // Ensure router is accessible in the template for button click
    };
    }
  }
  </script>
  
  <style scoped>
  .callback {
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    height: 100vh;
    font-size: 1.5rem;
  }
  </style>