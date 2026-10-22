import type { ComponentType } from 'react'
import type { IconProps } from '@/lib/phosphor'
import {
  ArrowsInSimple, Briefcase, Broom, Buildings, Bug, CalendarCheck, Camera, ChartLineUp, CheckCircle,
  ChefHat, Clock, CookingPot, DeviceMobile, DeviceMobileCamera, Drop, Factory, FileText, FirstAidKit,
  ForkKnife, GraduationCap, Handshake, HardHat, Heart, HouseLine, IdentificationBadge, Leaf, Lightning,
  MapPin, Medal, Path, PhoneCall, Plant, Recycle, ShieldCheck, Signature, Sparkle, Storefront, Truck,
  UserCircleCheck, UsersThree, Wrench,
} from '@/lib/phosphor'

// The icons an editor can pick in the CMS. This map must hold exactly the
// names in internal/domain/resource/defs.go (Icons); test/icons_test.go
// compares them, so every pickable icon renders.
export const ICONS: Record<string, ComponentType<IconProps>> = {
  'arrows-in-simple': ArrowsInSimple,
  'briefcase': Briefcase,
  'broom': Broom,
  'buildings': Buildings,
  'bug': Bug,
  'calendar-check': CalendarCheck,
  'camera': Camera,
  'chart-line-up': ChartLineUp,
  'check-circle': CheckCircle,
  'chef-hat': ChefHat,
  'clock': Clock,
  'cooking-pot': CookingPot,
  'device-mobile': DeviceMobile,
  'device-mobile-camera': DeviceMobileCamera,
  'drop': Drop,
  'factory': Factory,
  'file-text': FileText,
  'first-aid-kit': FirstAidKit,
  'fork-knife': ForkKnife,
  'graduation-cap': GraduationCap,
  'handshake': Handshake,
  'hard-hat': HardHat,
  'heart': Heart,
  'house-line': HouseLine,
  'identification-badge': IdentificationBadge,
  'leaf': Leaf,
  'lightning': Lightning,
  'map-pin': MapPin,
  'medal': Medal,
  'path': Path,
  'phone-call': PhoneCall,
  'plant': Plant,
  'recycle': Recycle,
  'shield-check': ShieldCheck,
  'signature': Signature,
  'sparkle': Sparkle,
  'storefront': Storefront,
  'truck': Truck,
  'user-circle-check': UserCircleCheck,
  'users-three': UsersThree,
  'wrench': Wrench,
}

export function Icon({ name, ...rest }: { name?: string } & IconProps) {
  const C = name ? ICONS[name] : undefined
  if (!C) return null
  return <C aria-hidden="true" {...rest} />
}
