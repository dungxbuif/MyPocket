import SwiftUI
import UIKit
import XCTest
@testable import MyPocketIOS

@MainActor
final class ReferenceDisplayTests: XCTestCase {
    func testReferenceSurfaceUsesItsNamedRadiusWithoutChangingOtherSurfaces() {
        XCTAssertEqual(MPSurfaceCardStyle.reference.cornerRadius, MPMetrics.referenceCornerRadius)
        XCTAssertEqual(MPSurfaceCardStyle.standard.cornerRadius, MPMetrics.cornerRadius)
        XCTAssertEqual(MPSurfaceCardStyle.elevated.cornerRadius, MPMetrics.cornerRadius)
    }

    func testMetricGridCompactsNormalMetricsAndExpandsForNarrowOrAccessibleContent() {
        let items = [
            MPMetricItem(title: "Số dư đầu tháng", value: "100.000 ₫"),
            MPMetricItem(title: "Số dư cuối kỳ", value: "5.100.000 ₫"),
            MPMetricItem(title: "Nạp ngoài chuyển ví", value: "0 ₫"),
            MPMetricItem(title: "Chuyển ví (thuần)", value: "5.000.000 ₫")
        ]
        let normal = size(MPMetricGrid(items: items), width: 343)
        let narrow = size(MPMetricGrid(items: items), width: 250)
        let accessible = size(MPMetricGrid(items: items).environment(\.dynamicTypeSize, .accessibility3), width: 343)
        XCTAssertEqual(normal.width, 343, accuracy: 1)
        XCTAssertLessThan(normal.height, 125, "Four ordinary metrics should occupy two compact rows.")
        XCTAssertGreaterThan(narrow.height, normal.height * 1.5)
        XCTAssertGreaterThan(accessible.height, narrow.height)
    }

    func testLongMetricLabelsAndMoneyGrowVerticallyWithinTheProposedWidth() {
        let long = MPMetricGrid(items: [
            MPMetricItem(title: "Số dư khi tạo ví mục tiêu trong kỳ báo cáo hiện tại", value: MPFormat.money(Int64.max)),
            MPMetricItem(title: "Còn lại", value: "0 ₫")
        ])
        let full = size(long, width: 343)
        let narrow = size(long, width: 150)
        XCTAssertEqual(full.width, 343, accuracy: 1)
        XCTAssertEqual(narrow.width, 150, accuracy: 1)
        XCTAssertGreaterThan(narrow.height, full.height)
        XCTAssertGreaterThan(narrow.height, 70, "The full value/label must expand, never be clipped to one line.")
    }

    func testMaskingKeepsMoneyOutOfMetricDisplayAndSpokenValue() {
        let metric = MPMetric(title: "Số dư", value: "9.223.372.036.854.775.807 ₫", masked: true, compact: true)
        XCTAssertEqual(metric.displayValue, MPFormat.maskedValue)
        XCTAssertEqual(metric.spokenValue, "Đã ẩn số tiền")
        let row = MPMetricRow(title: "2026-10-16", value: "120.000 ₫", masked: true)
        XCTAssertEqual(row.displayValue, MPFormat.maskedValue)
        XCTAssertEqual(row.spokenValue, "Đã ẩn số tiền")
    }

    func testCompactDailyRowFallsBackToStackingLongMoneyAndAccessibilityText() {
        let row = MPMetricRow(title: "2026-10-16", value: MPFormat.money(Int64.max))
        let full = size(row, width: 343)
        let narrow = size(row, width: 180)
        let accessible = size(row.environment(\.dynamicTypeSize, .accessibility3), width: 343)
        XCTAssertEqual(narrow.width, 180, accuracy: 1)
        XCTAssertGreaterThan(narrow.height, full.height)
        XCTAssertGreaterThan(accessible.height, full.height)
    }

    private func size<V: View>(_ view: V, width: CGFloat) -> CGSize {
        UIHostingController(rootView: view.environment(\.dynamicTypeSize, .large))
            .sizeThatFits(in: CGSize(width: width, height: 10_000))
    }
}
