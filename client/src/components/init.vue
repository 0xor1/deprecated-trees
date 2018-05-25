<template>
  <div class="c-c">
    <fingerprint-spinner :animation-duration="2000" :size="200" :color="color"></fingerprint-spinner>
  </div>
</template>

<script>
  import api from '@/api'
  import router from '@/router'
  import {FingerprintSpinner} from 'epic-spinners'
  export default {
    name: 'init',
    components: {FingerprintSpinner},
    data () {
      api.v1.centralAccount.getMe().then(() => {
        router.push('/me')
      }).catch(() => {
        router.push('/login')
      })
      return {
        color: 'white'
      }
    }
  }
</script>

<style scoped lang="scss">

</style>
