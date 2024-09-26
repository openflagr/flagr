<template>
    <div class="login-page flex flex-col items-center justify-center">
      <h1>Allen Flagger</h1>
      
      <div @click="loginWithGoogle" class="flex justify-center items-center btn-goggle">
        <img src="../../assets/images/google.svg" width="20" height="20" alt="My SVG Image" />
        <span>Login</span>
      </div>
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

<style lang="less" scoped>

.btn-goggle {
    padding: 2px;
    gap: 10px;
    width: 200px;
    background-color: rgb(226, 225, 225);
    color: black;
    border-radius: 50px;
    cursor: pointer;
    transition: all 0.3s;
    &:hover {
        background-color: rgb(159, 157, 157);
    }
}
/* Add some styling if needed */
</style>