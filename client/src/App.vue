<template>
  <div>
    <router-view></router-view>
  </div>
</template>

<script>
export default {
  name: 'app',
  data: {

  }
}
</script>

<style lang="scss">
  @import "var";
  html, body, div, .fill {
    @extend %fill;
    @extend %default;
  }
</style>
