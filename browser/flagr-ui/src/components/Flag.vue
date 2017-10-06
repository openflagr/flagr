<template>
  <div>
    <router-link :to="{name: 'home'}">Go back</router-link>
    <div v-if="loaded && flag">
      <h1>Flag {{ $route.params.id }}</h1>
      <div>
        <label>Description:</label> {{ flag.description }}
      </div>
    </div>
    <spinner v-if="!loaded"></spinner>
  </div>
</template>

<script>
import constants from '@/constants'
import fetchHelpers from '@/helpers/fetch'
import Spinner from '@/components/Spinner'

const {
  getJson
} = fetchHelpers

const {
  API_URL
} = constants

export default {
  name: 'flag',
  components: {
    spinner: Spinner
  },
  data () {
    return {
      loaded: false,
      flag: null
    }
  },
  created () {
    const flagId = this.$route.params.flagId
    getJson(`${API_URL}/flags/${flagId}`)
      .then(flag => {
        this.loaded = true
        this.flag = flag
      })
  }
}
</script>

<style scoped>
</style>
