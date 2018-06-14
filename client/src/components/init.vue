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
    name: 'init',
    components: {FingerprintSpinner},
    data () {
      api.v1.centralAccount.getMe().then((me) => {
        router.push('/app/region/' + me.region + '/shard/' + me.shard + '/account/' + me.id + '/projects')
      })
      return {}
    }
  }
</script>

<style scoped lang="scss">

</style>
