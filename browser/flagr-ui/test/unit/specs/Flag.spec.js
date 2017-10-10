import Vue from 'vue'
import Flag from '@/components/Flag'

describe('Flag.vue', () => {
  it('should construct vm correctly', () => {
    const Constructor = Vue.extend(Flag)
    const vm = new Constructor().$mount()
    expect(vm).to.not.be.a('null')
  })
})
