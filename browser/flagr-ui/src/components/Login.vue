<template>
    <div class="login-page">
      <h1>Login</h1>
      <button @click="loginWithGoogle">Login with Google</button>
    </div>
</template>
  
<script>
import { logout } from '../utils/apiUtil';
import { getGoogleRedirectionLink } from '../utils/apiUtil';
    import { mapState } from 'vuex';

    export default {
        methods: {
            async loginWithGoogle() {
                try {
                    // Call backend API to get the auth token and redirect URL
                    const authOrigin = window.location.origin
                    const data = {
                        uiRedirectUrl: `${authOrigin}/callback`,
                    }
                    getGoogleRedirectionLink(data)
                    .then((res) => {
                        console.log(res)
                        // console.log("test goo", res)
                        window.location.href = res.url
                    })
                    .catch((err) => {
                        console.log(err)
                        logout();
                    })
                } catch (error) {
                    console.error('Google login failed', error);
                }
            }
        },
        computed: {
            ...mapState({
                isAuth: (state) => state.isAuth, // Map userDetails from Vuex state
            }),
        },
        watch: {
            isAuth(newValue, oldValue) {
                console.log('isAuth changed:', oldValue, '->', newValue);
                // Add your logic here, e.g., redirect, update UI, etc.
            }
        },
        mounted() {
            console.log('isAuth:', this.isAuth);
            if(this.isAuth){
                this.$router.push('/');
            }
        }
    }
</script>

<style scoped>
/* Add some styling if needed */
</style>