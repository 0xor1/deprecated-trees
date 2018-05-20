<template>
  <div :class="{dark: useDark}">
    <router-view></router-view>
  </div>
</template>

<script>
export default {
  name: 'app',
  data () {
    return {
      useDark: true
    }
  }
}
</script>

<style lang="scss">
  @import "var";
  html, body, div, .fill {
    @extend %fill;
  }
  html, body, div, .def {
    @extend %default;
  }
  .center-children {
    justify-content: center;
    align-items: center;
  }
  .dark {
    @extend %darkTheme;
  }
</style>
