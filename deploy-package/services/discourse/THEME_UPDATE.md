# CERT Discourse Theme - Updated to Match cert-web

## 🎨 Overview

The Discourse forum theme has been completely updated to match the cert-web design system, creating a seamless visual experience across your entire platform.

## ✅ What's Been Updated

### Design System Alignment
- ✅ **Colors**: Exact match with cert-web (mint, electric, cyber)
- ✅ **Typography**: Inter font family (matching main site)
- ✅ **Spacing**: Consistent padding and margins
- ✅ **Border Radius**: Rounded corners matching cert-web (1rem, 0.75rem)
- ✅ **Shadows**: Subtle shadows for depth
- ✅ **Animations**: Smooth transitions and glow effects

### Components Styled

#### Core Elements
- ✅ Header with gradient background
- ✅ Navigation pills with hover states
- ✅ Topic list with card-style design
- ✅ Individual posts with borders and hover effects
- ✅ Sidebar sections with active states
- ✅ Search interface with focus states

#### Interactive Elements
- ✅ Primary buttons with mint/electric gradient
- ✅ Secondary buttons with subtle backgrounds
- ✅ Form inputs with focus glow
- ✅ Modals with rounded corners
- ✅ Dropdowns with dark theme
- ✅ Tags with electric blue styling

#### Content Formatting
- ✅ Code blocks with syntax highlighting
- ✅ Blockquotes with electric blue accent
- ✅ Links with mint hover color
- ✅ Headings with proper hierarchy
- ✅ Images with rounded corners

#### Special Features
- ✅ Cookie consent banner (matching main site)
- ✅ Google Analytics with consent mode
- ✅ Custom scrollbar (mint on hover)
- ✅ Pinned topics with mint highlight
- ✅ User profiles with avatar glow
- ✅ Notifications with unread states

### Responsive Design
- ✅ Mobile-optimized layouts
- ✅ Touch-friendly buttons
- ✅ Adaptive spacing
- ✅ Responsive modals

### Accessibility
- ✅ Focus-visible states with mint outline
- ✅ High contrast for notifications
- ✅ Keyboard navigation support
- ✅ Print-friendly styles

## 🎯 Color Palette

```scss
--cert-ink: #050508        // Main background
--cert-surface: #0A0A0F    // Elevated surfaces
--cert-panel: #0D0D14      // Cards and panels
--cert-mint: #00FFA3       // Primary accent
--cert-electric: #4D9FFF   // Secondary accent
--cert-cyber: #9D00FF      // Tertiary accent
--cert-text: #E4E4E7       // Primary text
--cert-text-muted: rgba(255, 255, 255, 0.6)  // Secondary text
```

## 📦 Files Modified

1. **common/common.scss** (763 lines)
   - Complete design system implementation
   - All component styles
   - Animations and effects
   - Responsive breakpoints

2. **common/header.html**
   - Google Analytics with consent mode
   - Cookie consent banner
   - Custom header navigation

3. **about.json**
   - Theme metadata
   - Color scheme definitions

## 🚀 Deployment Instructions

### Option 1: Upload Theme Package (Recommended)

1. **Download the theme package:**
   - Location: `cert-blockchain/deploy-package/services/discourse/cert-discourse-theme.tar.gz`

2. **Login to Discourse:**
   - URL: https://forum.c3rt.org/admin

3. **Navigate to Themes:**
   - Admin → Customize → Themes
   - URL: https://forum.c3rt.org/admin/customize/themes

4. **Install/Update Theme:**
   - If updating existing theme: Click "Edit CSS/HTML"
   - If new installation: Click "Install" → "Upload"
   - Replace the files with new content

5. **Activate Theme:**
   - Set as default theme
   - Preview to verify styling

### Option 2: Manual File Update

Copy the contents of these files to your Discourse theme:

1. **common/common.scss** → Theme CSS
2. **common/header.html** → Theme Header
3. **about.json** → Theme Settings

## 🔍 Verification Checklist

After deployment, verify these elements:

### Visual Consistency
- [ ] Background color matches cert-web (#050508)
- [ ] Buttons use mint/electric gradient
- [ ] Header has subtle gradient
- [ ] Cards have rounded corners (1rem)
- [ ] Borders are subtle (rgba(255, 255, 255, 0.1))

### Interactive Elements
- [ ] Hover states show mint color
- [ ] Focus states have mint outline
- [ ] Buttons have glow animation
- [ ] Links change color on hover
- [ ] Smooth transitions on all elements

### Typography
- [ ] Inter font is loaded and applied
- [ ] Text is readable (#E4E4E7)
- [ ] Headings are properly weighted
- [ ] Code blocks have proper styling

### Cookie Consent
- [ ] Banner appears on first visit
- [ ] Accept/Decline buttons work
- [ ] Choice is persisted
- [ ] Analytics respect consent

### Mobile Experience
- [ ] Layout adapts to small screens
- [ ] Touch targets are adequate
- [ ] Modals fit on screen
- [ ] Navigation is accessible

## 🎨 Design Features

### Animations
- **Glow Pulse**: Primary buttons pulse with mint glow
- **Smooth Transitions**: All interactive elements (0.2s ease)
- **Hover Effects**: Subtle color and transform changes

### Special Styling
- **Pinned Topics**: Mint background tint
- **Solved Topics**: Mint left border
- **Unread Notifications**: Mint badge
- **User Avatars**: Mint border with glow

## 📱 Browser Support

- ✅ Chrome/Edge (latest)
- ✅ Firefox (latest)
- ✅ Safari (latest)
- ✅ Mobile browsers (iOS/Android)

## 🆘 Troubleshooting

### Theme not applying
1. Clear browser cache (Ctrl+Shift+R)
2. Verify theme is set as default
3. Check for CSS errors in browser console

### Colors look different
1. Ensure color scheme is set to "CERT Dark"
2. Verify CSS variables are loaded
3. Check for conflicting plugins

### Fonts not loading
1. Verify Google Fonts import in common.scss
2. Check network tab for font loading
3. Clear CDN cache if applicable

## 📊 Performance

- **CSS Size**: ~25KB (uncompressed)
- **Load Time**: Minimal impact
- **Animations**: GPU-accelerated
- **Mobile**: Optimized for performance

## 🔄 Future Updates

To keep the theme in sync with cert-web:
1. Monitor cert-web design changes
2. Update color variables as needed
3. Add new component styles
4. Test across devices

## 📞 Support

For issues or questions:
- Check Discourse admin logs
- Review browser console for errors
- Test in incognito mode
- Compare with cert-web styling

---

**Theme Version**: 2.0
**Last Updated**: 2026-01-09
**Compatible with**: Discourse 3.x

