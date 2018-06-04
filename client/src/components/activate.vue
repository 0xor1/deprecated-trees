<template>
  <v-app id="inspire">
    <v-content>
      <v-container fluid fill-height style="background-color: #9FC657;">
        <v-layout align-center justify-center>
          <fingerprint-spinner :animation-duration="2000" :size="200" :color="'#405A0F'"></fingerprint-spinner>
        </v-layout>
      </v-container>
    </v-content>
  </v-app>
</template>

<script>
  import api from '@/api'
  import router from '@/router'
  import {FingerprintSpinner} from 'epic-spinners'
  export default {
    name: 'activate',
    components: {FingerprintSpinner},
    data () {
      api.v1.centralAccount.activate(router.currentRoute.query.email, router.currentRoute.params.activationCode).then(() => {
        router.push('/login')
      }).catch(() => {
        // TODO
      })
      return {}
    }
  }
</script>

<style scoped lang="scss">

</style>
