<template>
  <div class="center-children">
    <fingerprint-spinner :animation-duration="2000" :size="200" :color="'white'"></fingerprint-spinner>
  </div>
</template>

<script>
  import {FingerprintSpinner} from 'epic-spinners'
  export default {
    name: 'init',
    components: {FingerprintSpinner},
    data () {
      return {}
    }
  }
</script>

<style scoped lang="scss">

</style>
