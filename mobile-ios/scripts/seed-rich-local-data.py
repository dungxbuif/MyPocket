#!/usr/bin/env python3
"""Add deterministic, resumable manual-UAT data to an existing LOCAL account.

Uses application APIs, never resets data. Token comes from a private file.
Month configuration backfill uses local SQL for newly created sample jars:
the API can only edit configs already present in a materialized month.
The journal must be kept when resuming: some legacy routes lack idempotency.
An interrupted request on those routes needs reconciliation before rerunning.
"""
import argparse
import calendar
import datetime as dt
from decimal import Decimal, ROUND_HALF_UP
import hashlib
import json
from pathlib import Path
import random
import subprocess
import urllib.error
import urllib.parse
import urllib.request
import uuid
from zoneinfo import ZoneInfo


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--base-url", default="http://127.0.0.1:8081/api/v1")
    parser.add_argument("--token-file", required=True)
    parser.add_argument("--owner-id", required=True)
    parser.add_argument("--journal", required=True)
    parser.add_argument("--local-db", required=True)
    parser.add_argument("--through", default="2026-11-02")
    args = parser.parse_args()
    if not args.local_db.startswith("mypocket_native_ui_") or not all(c.isalnum() or c in "_-" for c in args.local_db):
        raise SystemExit("Only a native UI local fixture database is allowed")
    uuid.UUID(args.owner_id)
    url = urllib.parse.urlparse(args.base_url)
    if url.scheme != "http" or url.hostname not in {"127.0.0.1", "localhost", "::1"} or url.username:
        raise SystemExit("Only loopback HTTP is allowed")
    today = dt.datetime.now(ZoneInfo("Asia/Ho_Chi_Minh")).date()
    through = dt.date.fromisoformat(args.through)
    if through != dt.date(2026, 11, 2) or through > today:
        raise SystemExit("This fixture is designed for 2026-11-02; future runs need a new dataset/journal")
    token = Path(args.token_file).read_text().strip()
    journal_path = Path(args.journal)
    identity = {"owner": args.owner_id, "base_url": args.base_url, "through": args.through}
    journal = json.loads(journal_path.read_text()) if journal_path.exists() else {"identity": identity, "writes": {}}
    if journal["identity"] != identity:
        raise SystemExit("Journal belongs to a different seed target")
    if journal.get("pending"):
        raise SystemExit("Reconcile the journal pending request before resuming")

    def persist():
        tmp = journal_path.with_suffix(".tmp")
        tmp.write_text(json.dumps(journal, ensure_ascii=False, indent=2))
        tmp.chmod(0o600)
        tmp.replace(journal_path)

    def request(method, path, data=None, key=None):
        headers = {"Authorization": "Bearer " + token, "Content-Type": "application/json"}
        if key:
            headers["Idempotency-Key"] = "rich2026-" + hashlib.sha256(key.encode()).hexdigest()[:40]
        req = urllib.request.Request(args.base_url + path, method=method, headers=headers,
                                     data=None if data is None else json.dumps(data).encode())
        try:
            with urllib.request.urlopen(req, timeout=45) as response:
                body = response.read()
                return json.loads(body).get("data") if body else None
        except urllib.error.HTTPError as error:
            detail = error.read().decode()
            raise RuntimeError(f"{method} {path}: HTTP {error.code}: {detail}") from None

    def write(key, path, data, method="POST"):
        digest = hashlib.sha256(json.dumps([method, path, data], sort_keys=True).encode()).hexdigest()
        if key in journal["writes"]:
            previous = journal["writes"][key]
            if previous["digest"] != digest:
                raise RuntimeError("Seed payload changed for " + key)
            return previous["result"]
        journal["pending"] = {"key": key, "path": path, "method": method, "data": data}
        persist()
        result = request(method, path, data, key)
        journal["writes"][key] = {"digest": digest, "result": result}
        journal.pop("pending")
        persist()
        if len(journal["writes"]) % 100 == 0:
            print(f"Persisted {len(journal['writes'])} API operations", flush=True)
        return result

    profile = request("GET", "/auth/profile")
    if profile["id"] != args.owner_id or profile["email"] != "native-fixture@example.test":
        raise SystemExit("Refusing unexpected account; this seed targets the native local fixture")
    if profile["timezone"] != "Asia/Ho_Chi_Minh":
        raise SystemExit("Account must already use Asia/Ho_Chi_Minh")
    print("Local account verified:", profile["email"], flush=True)
    categories = {x["icon_key"]: x["id"] for x in request("GET", "/categories") if "• Mẫu" not in x["name"]}
    wallets = {}
    definitions = [
        ("bank", "Ngân hàng • Mẫu 2026", "basic", 120000000, {}),
        ("cash", "Tiền mặt • Mẫu 2026", "basic", 5000000, {}),
        ("online", "Ví mua sắm • Mẫu 2026", "basic", 3000000, {}),
        ("goal", "Mua xe • Mẫu 2026", "goal", 0, {"target_amount": 80000000, "target_date": "2027-06-30", "goal_notifications_enabled": True}),
        ("done", "Laptop đã đủ • Mẫu 2026", "goal", 0, {"target_amount": 25000000, "target_date": "2026-10-31", "goal_notifications_enabled": True}),
        ("late", "Du lịch còn thiếu • Mẫu 2026", "goal", 0, {"target_amount": 30000000, "target_date": "2026-09-30"}),
        ("credit", "Thẻ mua sắm • Mẫu 2026", "credit", 0, {"credit_limit": 50000000, "statement_day": 20, "payment_due_day": 5, "last_statement_balance": 3500000}),
    ]
    for key, name, kind, opening, extra in definitions:
        wallets[key] = write("wallet-" + key, "/wallets", dict(name=name, type=kind, opening_balance=opening,
            is_in_total=kind != "credit", description="Dữ liệu mẫu để kiểm tra sản phẩm năm 2026", **extra))["id"]

    # Stable jars with configurations for every month, including a historical-only jar.
    jars = {}
    jar_defs = [("food", "Ăn uống • Mẫu", 5500000), ("home", "Nhà cửa • Mẫu", 9000000),
                ("learn", "Học tập • Mẫu", 1800000), ("fun", "Tận hưởng • Mẫu", 2200000),
                ("health", "Sức khỏe • Mẫu", 1500000), ("tet", "Tết 2026 • Mẫu", 8000000)]
    for key, name, amount in jar_defs:
        result = write("jar-" + key, "/jars", dict(month="2026-01", name=name, allocation_mode="fixed", allocation_amount=amount))
        jars[key] = result["jar_id"]
        uuid.UUID(jars[key])
    # Safe additive backfill limited to these new fixture-owned jars. Existing
    # month configurations and manually entered content are never replaced.
    jar_ids = ",".join("'" + value + "'" for value in jars.values())
    count = subprocess.check_output(["psql", "-X", "-h", "127.0.0.1", "-d", args.local_db, "-Atc",
        f"SELECT count(*) FROM jars WHERE owner_id='{args.owner_id}' AND id IN ({jar_ids})"], text=True).strip()
    if count != str(len(jars)):
        raise SystemExit("Local database does not contain the exact API-created jars")
    sql = f"""BEGIN;
    INSERT INTO jar_months(owner_id,month)
      SELECT '{args.owner_id}', d::date FROM generate_series('2026-01-01'::date,
      '{through.replace(day=1)}'::date, interval '1 month') d ON CONFLICT DO NOTHING;
    INSERT INTO jar_month_configs(owner_id,month,jar_id,name,allocation_mode,allocation_amount,allocation_percent_bps,active)
      SELECT c.owner_id,d::date,c.jar_id,c.name,c.allocation_mode,c.allocation_amount,c.allocation_percent_bps,true
      FROM jar_month_configs c CROSS JOIN generate_series('2026-02-01'::date,
      '{through.replace(day=1)}'::date,interval '1 month') d
      WHERE c.owner_id='{args.owner_id}' AND c.month='2026-01-01' AND c.jar_id IN ({jar_ids})
      AND (c.jar_id<>'{jars['tet']}' OR d<'2026-03-01') ON CONFLICT DO NOTHING;
    COMMIT;"""
    subprocess.run(["psql", "-X", "-h", "127.0.0.1", "-d", args.local_db, "-v", "ON_ERROR_STOP=1", "-q"], input=sql, text=True, check=True)
    for key, name, amount in jar_defs:
        for month in range(2, through.month + 1):
            if key == "tet" and month > 2:
                continue
            period = f"2026-{month:02}"
            write(f"jar-{key}-{period}", f"/jars/{jars[key]}/months/{period}",
                  dict(name=name, allocation_mode="fixed", allocation_amount=amount), "PUT")

    def date(month, day, hour=12):
        return f"2026-{month:02}-{day:02}T{hour:02}:15:00+07:00"

    def tx(key, wallet, icon, amount, stamp, note, jar=None, included=True, travel=None):
        day = dt.datetime.fromisoformat(stamp).date()
        if day > through:
            return None
        payload = dict(wallet_id=wallets.get(wallet, wallet), category_id=categories[icon],
                       type="income" if icon.startswith("income_") else "expense", amount=amount,
                       occurred_at=stamp, note=note + " [Mẫu 2026]", included_in_reports=included, auto_link_travel=False)
        if jar:
            payload["jar_id"] = jars[jar]
        if travel:
            payload["travel_event_id"] = travel
        return write("tx-" + key, "/transactions", payload)

    def transfer(key, source, destination, amount, stamp, note):
        if dt.datetime.fromisoformat(stamp).date() <= through:
            return write("transfer-" + key, "/transactions/transfer", dict(source_wallet_id=wallets.get(source, source),
                destination_wallet_id=wallets.get(destination, destination), amount=amount, occurred_at=stamp, note=note + " [Mẫu 2026]"))

    rng = random.Random(20261102)
    expenses = [
        ("expense_food_meals", 45000, 95000, "Cơm trưa văn phòng", "food"),
        ("expense_food_coffee", 25000, 65000, "Cà phê cùng đồng nghiệp", "food"),
        ("expense_food_snacks", 15000, 75000, "Ăn vặt cuối giờ", "food"),
        ("expense_fuel", 70000, 180000, "Đổ xăng", None),
        ("expense_shopping_household", 150000, 750000, "Đi siêu thị mua đồ gia dụng", "home"),
        ("expense_books", 90000, 320000, "Mua sách", "learn"),
        ("expense_movies", 100000, 260000, "Xem phim cuối tuần", "fun"),
        ("expense_medicine", 45000, 280000, "Mua thuốc và vitamin", "health"),
        ("expense_family_pets", 80000, 380000, "Thức ăn và chăm sóc mèo", None),
        ("expense_taxi", 35000, 140000, "Di chuyển bằng taxi", None),
        ("expense_clothes", 180000, 650000, "Mua quần áo", "fun"),
        ("expense_restaurant", 250000, 800000, "Ăn tối gia đình", "food"),
    ]
    day = dt.date(2026, 1, 1)
    while day <= through:
        # Two to four varied entries each day; monthly differences aid comparison charts.
        count = 2 + (day.day % 3)
        for item in range(count):
            icon, low, high, note, jar = expenses[(day.toordinal() + item * 5) % len(expenses)]
            amount = rng.randrange(low // 1000, high // 1000 + 1) * 1000
            if day.month in (2, 7):
                amount = int(amount * 1.4 / 1000) * 1000
            wallet = ["bank", "cash", "online"][(day.day + item) % 3]
            tx(f"daily-{day}-{item}", wallet, icon, amount, date(day.month, day.day, 8 + item * 3), note, jar)
        day += dt.timedelta(days=1)

    for month in range(1, through.month + 1):
        period = f"2026-{month:02}"
        tx(f"salary-{month}", "bank", "income_salary", 42000000 if month < 7 else 48000000, date(month, 1, 7), f"Lương tháng {month}")
        tx(f"freelance-{month}", "bank", "income_other", (month % 3 + 1) * 2200000, date(month, 18), "Thiết kế tự do — thanh toán dự án")
        tx(f"rent-{month}", "bank", "expense_bills_rent", 6500000, date(month, 2), "Tiền thuê nhà", "home")
        tx(f"electricity-{month}", "bank", "expense_bills_electricity", 650000 + month % 4 * 170000, date(month, 8), "Hóa đơn điện", "home")
        tx(f"internet-{month}", "bank", "expense_bills_internet", 280000, date(month, 10), "Internet gia đình", "home")
        tx(f"education-{month}", "bank", "expense_education", 1200000, date(month, 12), "Khóa học ngoại ngữ", "learn")
        tx(f"gift-{month}", "bank", "income_gift", 500000, date(month, 20), "Quà sinh nhật / gia đình")
        tx(f"exclude-{month}", "bank", "expense_business", 400000, date(month, 15), "Chi hộ công ty — không tính báo cáo", included=False)
        transfer(f"cash-{month}", "bank", "cash", 12000000, date(month, 1), "Rút tiền mặt chi tiêu")
        transfer(f"online-{month}", "bank", "online", 10000000, date(month, 1), "Nạp ví mua sắm")
        transfer(f"goal-{month}", "bank", "goal", 4000000, date(month, 3), "Góp mục tiêu mua xe")
        transfer(f"done-{month}", "bank", "done", 2500000, date(month, 4), "Góp mua laptop")
        if month <= 3:
            transfer(f"late-{month}", "bank", "late", 2000000, date(month, 9), "Góp du lịch")
        if month < through.month:
            tx(f"goal-interest-{month}", "goal", "income_interest", 40000 + month * 2000, date(month, 28), "Lãi tiết kiệm mục tiêu")
        # Actual bank card purchase/payment/refund semantics, not ordinary transactions.
        for kind, amount, day_number, cat, note in [
            ("purchase", 2200000, 6, "expense_shopping", "Thanh toán thẻ mua sắm"),
            ("fee", 55000, 19, "expense_cost", "Phí dịch vụ thẻ"),
            ("refund", 200000, 22, "income_other", "Hoàn tiền món hàng trả lại"),
        ]:
            if dt.date(2026, month, day_number) <= through:
                write(f"credit-{month}-{kind}", f"/credit/wallets/{wallets['credit']}/entries", dict(kind=kind,
                    amount=amount, category_id=categories[cat], occurred_at=date(month, day_number), note=note + " [Mẫu 2026]"))
        if dt.date(2026, month, 25) <= through:
            write(f"credit-pay-{month}", f"/credit/wallets/{wallets['credit']}/payments", dict(source_wallet_id=wallets["bank"],
                amount=1800000, occurred_at=date(month, 25), note="Thanh toán một phần dư nợ [Mẫu 2026]"))
        for kind, icon, limit in [("food", "expense_food", 5000000), ("shopping", "expense_shopping", 1500000)]:
            write(f"budget-{kind}-{month}", "/budgets", dict(name=f"{('Ăn uống' if kind == 'food' else 'Mua sắm')} T{month} • Mẫu",
                limit_amount=limit, category_id=categories[icon], start_date=f"{period}-01",
                end_date=f"{period}-{calendar.monthrange(2026, month)[1]}"))
        # Do not overwrite a month note the user already wrote.
        existing_month = request("GET", "/months/" + period)
        if not existing_month.get("note"):
            write("note-" + period, f"/months/{period}/note", {"note": f"[Mẫu 2026] Tháng {month}: theo dõi chi phí sinh hoạt, khoản tiết kiệm và mua sắm; xem biểu đồ để so sánh tháng trước."}, "PUT")
        print("Prepared month", period, flush=True)

    for month, icon, amount, note, jar in [
        (2, "expense_gifts", 8000000, "Lì xì và quà Tết", "tet"),
        (2, "income_bonus", 30000000, "Thưởng Tết", None),
        (4, "expense_electronics", 18500000, "Mua máy ảnh phục vụ công việc", "fun"),
        (7, "expense_medical_treatment", 12500000, "Chi phí điều trị bất thường", "health"),
        (10, "income_sell_items", 3200000, "Bán điện thoại cũ", None),
    ]:
        tx(f"event-{month}-{icon}", "bank", icon, amount, date(month, 16), note, jar)
    transfer("goal-withdraw", "goal", "bank", 1500000, date(9, 15), "Rút một phần mục tiêu sửa xe")
    # One current month near-limit/over-limit example without adding future ledger entries.
    tx("nov-shopping", "online", "expense_shopping", 1850000, date(11, 1), "Mua máy hút bụi — ngân sách mua sắm vượt mức", "fun")
    write("budget-series-home", "/budget-series", dict(name="Nhà cửa hàng tháng • Mẫu", limit_amount=9000000,
        category_id=categories["expense_bills"], carryover_enabled=True, cadence="monthly", start_date="2026-11-01"))
    write("budget-series-week", "/budget-series", dict(name="Ăn uống hàng tuần • Mẫu", limit_amount=1300000,
        wallet_id=wallets["bank"], category_id=categories["expense_food"], carryover_enabled=False, cadence="weekly", start_date="2026-11-02"))

    for key, name, icon, amount, frequency, active, next_run in [
        ("rent", "Tiền thuê nhà • Mẫu", "expense_bills_rent", 6500000, "monthly", True, "2026-12-02T08:00:00+07:00"),
        ("salary", "Lương hàng tháng • Mẫu", "income_salary", 48000000, "monthly", True, "2026-12-01T08:00:00+07:00"),
        ("net", "Internet • Mẫu", "expense_bills_internet", 280000, "monthly", True, "2026-11-10T08:00:00+07:00"),
        ("gym", "Thẻ tập đang tạm dừng • Mẫu", "expense_health_fitness", 650000, "monthly", False, "2026-11-05T08:00:00+07:00"),
        ("coffee", "Cà phê tuần • Mẫu", "expense_food_coffee", 120000, "weekly", True, "2026-11-06T08:00:00+07:00"),
    ]:
        write("recurring-" + key, "/recurring", dict(name=name, wallet_id=wallets["bank"], category_id=categories[icon],
            type="income" if icon.startswith("income_") else "expense", amount=amount, frequency=frequency, interval=1,
            next_run_at=next_run, active=active, note="Lịch mẫu để thử sửa, tạm dừng và tiếp tục"))

    for key, name, start, end, month, days in [
        ("dalat", "Đà Lạt tháng 3 • Mẫu", "2026-03-20", "2026-03-23", 3, [20, 21, 22]),
        ("danang", "Đà Nẵng hè • Mẫu", "2026-07-10", "2026-07-14", 7, [10, 11, 12, 13]),
        ("hanoi", "Công tác Hà Nội • Mẫu", "2026-11-01", "2026-11-04", 11, [1, 2]),
    ]:
        trip = write("travel-" + key, "/travel/events", dict(name=name, context="Dữ liệu mẫu: vé, khách sạn, ăn uống, di chuyển",
            starts_on=start, ends_on=end, active=False))["id"]
        for day_number in days:
            tx(f"trip-{key}-{day_number}", "bank", "expense_travel", 650000 + day_number % 3 * 250000,
                date(month, day_number), "Khách sạn và vé tham quan " + name, "fun", travel=trip)
            tx(f"trip-food-{key}-{day_number}", "cash", "expense_restaurant", 320000,
                date(month, day_number, 18), "Ăn tối chuyến đi " + name, "food", travel=trip)

    catalog = request("GET", "/portfolio/catalog-config")
    asset_defs = [
        ("chi", "Vàng nhẫn tích lũy • Mẫu", "chỉ", "manual", None, 9500000, 7800000),
        ("luong", "Vàng miếng mua cao • Mẫu", "lượng", "manual", None, 90000000, 98000000),
        ("phan", "Vàng lẻ từng phân • Mẫu", "phân", "manual", None, 940000, 810000),
        ("unknown", "Vàng chưa định giá • Mẫu", "chỉ", "manual", None, None, 8500000),
        ("closed", "Vàng đã bán hết • Mẫu", "chỉ", "manual", None, 9600000, 8300000),
    ]
    for i, product in enumerate(catalog["products"]):
        asset_defs.append((f"auto{i}", product["name"] + " • Mẫu", product["unit"], "automatic", product["id"], None, 8200000 + i * 200000))
    for key, name, unit, mode, product, price, cost in asset_defs:
        payload = dict(symbol="DEMO-" + key.upper(), name=name, unit_label=unit, price_mode=mode)
        if product:
            payload["product_id"] = product
        asset = write("asset-" + key, "/portfolio/assets", payload)["id"]
        quantities = ["0.2", "0.3", "0.1"] if unit == "lượng" else ["2", "1.5", "0.5"]
        for i, month in enumerate([1, 4, 8]):
            write(f"trade-{key}-{i}", f"/portfolio/assets/{asset}/trades", dict(side="buy", quantity=quantities[i],
                unit_price=cost + i * (1000000 if unit == "lượng" else 10000), fee=25000,
                occurred_at=date(month, 15), note="Tích lũy theo đợt [Mẫu 2026]"))
        sell_qty = "4" if key == "closed" else ("0.1" if unit == "lượng" else "0.5")
        write("sell-" + key, f"/portfolio/assets/{asset}/trades", dict(side="sell", quantity=sell_qty, unit_price=int(cost * 1.12),
            fee=10000, occurred_at=date(10, 20), note="Bán một phần / chốt vị thế [Mẫu 2026]"))
        if mode == "manual" and price is not None:
            write("price-" + key, f"/portfolio/assets/{asset}/price", {"latest_price": price}, "PATCH")

    # Supporting screens: custom categories, archived history, adjustments,
    # credit interest, feedback, and mixed read/unread notification states.
    parent = write("category-parent", "/categories", dict(name="Dự án cá nhân • Mẫu", kind="expense", icon_key="expense_business"))["id"]
    child = write("category-child", "/categories", dict(name="Thiết bị dự án • Mẫu", kind="expense", icon_key="expense_electronics",
        parent_id=parent, wallet_ids=[wallets["bank"]]))["id"]
    archived = write("category-archived", "/categories", dict(name="Khóa học đã kết thúc • Mẫu", kind="expense", icon_key="expense_education"))["id"]
    write("category-income", "/categories", dict(name="Nhuận bút • Mẫu", kind="income", icon_key="income_other"))
    write("category-debt", "/categories", dict(name="Cho bạn mượn • Mẫu", kind="debt", icon_key="debt_lend"))
    for key, cat, amount, stamp in [("custom", child, 1250000, date(5, 8)), ("archived", archived, 1800000, date(3, 9))]:
        write("tx-category-" + key, "/transactions", dict(wallet_id=wallets["bank"], category_id=cat, type="expense", amount=amount,
            occurred_at=stamp, note="Lịch sử danh mục tùy chỉnh [Mẫu 2026]", included_in_reports=True, auto_link_travel=False))
    write("archive-category", f"/categories/{archived}/status", {"is_active": False}, "PATCH")
    for direction, amount in [("increase", 85000), ("decrease", 35000)]:
        write("adjustment-" + direction, "/transactions/adjustment", dict(wallet_id=wallets["cash"], direction=direction,
            amount=amount, occurred_at=date(6, 30), note="Đối chiếu tiền mặt [Mẫu 2026]"))
    write("credit-interest", f"/credit/wallets/{wallets['credit']}/entries", dict(kind="interest", amount=125000,
        category_id=categories["expense_interest_paid"], occurred_at=date(10, 28), note="Lãi dư nợ mẫu"))
    for kind, title in [("bug", "Kiểm tra nội dung dài trong phản hồi"), ("feature", "Gợi ý mục tiêu tiết kiệm"), ("improvement", "Kiểm tra trạng thái đã giải quyết")]:
        feedback = write("feedback-" + kind, "/feedback", dict(type=kind, title=title + " • Mẫu",
            description="Đây là phản hồi mẫu phục vụ kiểm tra danh sách, chi tiết và thao tác mở lại; không phải báo lỗi sản phẩm thật."))
        if kind == "improvement":
            write("feedback-resolved", f"/feedback/{feedback['id']}/resolution", {"resolved": True}, "PATCH")
    notifications = request("GET", "/notifications")
    sample_notifications = [x for x in notifications if "• Mẫu" in x["title"] and "80%" in x["body"]]
    if sample_notifications and "notification-read" not in journal["writes"]:
        write("notification-read", f"/notifications/{sample_notifications[0]['id']}/read", {})
    write("profile-timezone", "/auth/profile", {"timezone": "Asia/Ho_Chi_Minh", "initialize_only": True}, "PATCH")

    # Portfolio's API deliberately does not move wallet money. Add separately
    # reconciled cashflows so fixtures do not count purchase funding twice.
    wallets["goldcash"] = write("wallet-goldcash", "/wallets", dict(name="Vốn mua vàng • Mẫu 2026", type="basic",
        opening_balance=300000000, is_in_total=True, description="Vốn đầu năm; mọi lần mua/bán vàng mẫu có dòng tiền đối ứng"))["id"]
    for key, item in list(journal["writes"].items()):
        if not (key.startswith("trade-") or key.startswith("sell-")):
            continue
        trade = item["result"]
        gross = int((Decimal(trade["quantity"]) * trade["unit_price"]).quantize(Decimal(1), rounding=ROUND_HALF_UP))
        buying = trade["side"] == "buy"
        tx("funding-" + key, "goldcash", "expense_investment" if buying else "income_sell_items",
           gross + trade["fee"] if buying else gross - trade["fee"], trade["occurred_at"],
           ("Vốn mua vàng " if buying else "Tiền bán vàng ") + key + "; tách khỏi chi tiêu sinh hoạt", included=False)
    write("jar-percent-current", f"/jars/{jars['learn']}/months/2026-11",
          dict(name="Học tập • Mẫu", allocation_mode="percent", allocation_percent=5), "PUT")
    write("jar-none-current", f"/jars/{jars['health']}/months/2026-11",
          dict(name="Sức khỏe • Mẫu", allocation_mode="none"), "PUT")
    trips = request("GET", "/travel/events")
    if not any(x["active"] for x in trips):
        trip_id = journal["writes"]["travel-hanoi"]["result"]["id"]
        write("travel-active", f"/travel/events/{trip_id}/activate", {})
    closed_card = write("wallet-credit-paid", "/wallets", dict(name="Thẻ đã tất toán • Mẫu", type="credit", opening_balance=0,
        is_in_total=False, credit_limit=10000000, last_statement_balance=0, statement_day=20, payment_due_day=5))["id"]
    write("credit-paid-purchase", f"/credit/wallets/{closed_card}/entries", dict(kind="purchase", amount=1000000,
        category_id=categories["expense_shopping"], occurred_at=date(8, 10), note="Khoản thẻ đã thanh toán đủ [Mẫu 2026]"))
    write("credit-paid-payment", f"/credit/wallets/{closed_card}/payments", dict(source_wallet_id=wallets["bank"], amount=1000000,
        occurred_at=date(8, 25), note="Tất toán thẻ [Mẫu 2026]"))
    write("recurring-ended", "/recurring", dict(name="Lịch hai ngày đã kết thúc • Mẫu", wallet_id=wallets["bank"],
        category_id=categories["expense_food_coffee"], type="expense", amount=30000, frequency="daily", interval=1,
        next_run_at="2026-11-01T08:00:00+07:00", ends_at="2026-11-02T08:00:00+07:00", active=True,
        note="Giao dịch được sinh từ lịch định kỳ [Mẫu 2026]"))
    write("recurring-materialize", "/recurring/run-due", {})

    journal["completed"] = True
    persist()
    print(json.dumps({"operations": len(journal["writes"]), "through": str(through), "journal": str(journal_path)}, ensure_ascii=False))


if __name__ == "__main__":
    main()
