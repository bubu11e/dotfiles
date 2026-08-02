import AppKit
import Foundation

// Rasterise an emoji into the PWA icon set. AppKit is used rather than a
// third-party rasteriser because it is the only emoji-capable renderer present on
// a stock macOS, and CI never runs this: the PNGs are committed once.

struct Args {
    var emoji = "\u{1F680}"
    var outDir = "."
    var background = "#ffffff"
}

func parse() -> Args {
    var a = Args()
    var it = CommandLine.arguments.dropFirst().makeIterator()
    while let flag = it.next() {
        switch flag {
        case "--emoji": a.emoji = it.next() ?? a.emoji
        case "--out": a.outDir = it.next() ?? a.outDir
        case "--background": a.background = it.next() ?? a.background
        default: break
        }
    }
    return a
}

func color(_ hex: String) -> NSColor {
    var s = hex.trimmingCharacters(in: CharacterSet(charactersIn: "#"))
    if s.count == 3 { s = s.map { "\($0)\($0)" }.joined() }
    guard let v = UInt32(s, radix: 16), s.count == 6 else { return .white }
    return NSColor(srgbRed: CGFloat((v >> 16) & 0xff) / 255,
                   green: CGFloat((v >> 8) & 0xff) / 255,
                   blue: CGFloat(v & 0xff) / 255, alpha: 1)
}

func render(emoji: String, size: Int, background: NSColor?, inset: CGFloat) -> Data? {
    let px = CGFloat(size)
    guard let rep = NSBitmapImageRep(bitmapDataPlanes: nil, pixelsWide: size, pixelsHigh: size,
                                     bitsPerSample: 8, samplesPerPixel: 4, hasAlpha: true,
                                     isPlanar: false, colorSpaceName: .deviceRGB,
                                     bytesPerRow: 0, bitsPerPixel: 0) else { return nil }
    NSGraphicsContext.saveGraphicsState()
    NSGraphicsContext.current = NSGraphicsContext(bitmapImageRep: rep)
    let full = NSRect(x: 0, y: 0, width: px, height: px)
    if let bg = background {
        bg.setFill()
        full.fill()
    }
    let glyph = px * (1 - 2 * inset)
    let font = NSFont(name: "Apple Color Emoji", size: glyph) ?? NSFont.systemFont(ofSize: glyph)
    let attrs: [NSAttributedString.Key: Any] = [.font: font]
    let str = NSAttributedString(string: emoji, attributes: attrs)
    let bounds = str.size()
    str.draw(at: NSPoint(x: (px - bounds.width) / 2, y: (px - bounds.height) / 2))
    NSGraphicsContext.restoreGraphicsState()
    return rep.representation(using: .png, properties: [:])
}

let args = parse()
let bg = color(args.background)
let outputs: [(String, Int, NSColor?, CGFloat)] = [
    ("icon-192.png", 192, bg, 0.12),
    ("icon-512.png", 512, bg, 0.12),
    ("icon-maskable-512.png", 512, bg, 0.22),
    ("apple-touch-icon.png", 180, bg, 0.10),
]
for (name, size, background, inset) in outputs {
    guard let data = render(emoji: args.emoji, size: size, background: background, inset: inset) else {
        FileHandle.standardError.write("failed to render \(name)\n".data(using: .utf8)!)
        exit(1)
    }
    let url = URL(fileURLWithPath: args.outDir).appendingPathComponent(name)
    try data.write(to: url)
    print("wrote \(url.path)")
}
